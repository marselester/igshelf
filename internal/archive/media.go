// Package archive provides access to Instagram account data downloaded from https://www.instagram.com/download/request/.
// The archive contains JSON and media files which are organized in directories by year/month.
package archive

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/marselester/igshelf"
)

// tocFilename is a JSON file that describes archived media files (table of contents).
const tocFilename = "media.json"

// NewService creates a media service that provides access to Instagram timeline from zip archive.
// It opens an archive and maps paths to corresponding media files.
func NewService(filename string) (*MediaService, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	s := MediaService{
		r:   r,
		toc: make(map[string]*zip.File, len(r.File)),
	}
	for _, f := range r.File {
		if f.Name == tocFilename || strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".mp4") {
			s.toc[f.Name] = f
		}
	}

	return &s, nil
}

// MediaService represents a service to work with an Instagram archive.
type MediaService struct {
	r *zip.ReadCloser
	// toc maps paths to corresponding media files in archive r.
	toc map[string]*zip.File
}

// Close closes the underlying zip file.
func (s *MediaService) Close() error {
	return s.r.Close()
}

// Download copies the media file from its location in archive.
// Note, thumbnail is not available.
func (s *MediaService) Download(ctx context.Context, m *igshelf.Media) (content, thumbnail []byte, err error) {
	f, ok := s.toc[m.Location]
	if !ok {
		return nil, nil, fmt.Errorf("file not found in archive %s", m.Location)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file in archive %s: %w", m.Location, err)
	}
	defer rc.Close()

	if content, err = ioutil.ReadAll(rc); err != nil {
		return nil, nil, fmt.Errorf("failed to read content: %w", err)
	}
	return content, nil, nil
}

// media represents an image or video (album is not available).
type media struct {
	// Caption is the media's caption text, e.g., "Still jumping".
	Caption string `json:"caption"`
	// TakenAt is the media's publish date, e.g., 2020-10-07T15:55:33+00:00.
	TakenAt time.Time `json:"taken_at"`
	// Path is a relative path to a media file, e.g., videos/202010/8c996aa535f0f7a322d4dbaef9cfd266.mp4.
	Path string `json:"path"`
}

// nomenclature represents content of media.json found in a zip archive.
type nomenclature struct {
	Videos []*media `json:"videos"`
	Photos []*media `json:"photos"`
}

// List returns a collection of media in reverse chronological order (newest first).
// Note, media is not sorted in zip archive, so the order is restored based on date and caption.
func (s *MediaService) List(ctx context.Context) igshelf.MediaIter {
	iter := MediaIter{
		ctx: ctx,
	}
	f, ok := s.toc[tocFilename]
	if !ok {
		iter.err = fmt.Errorf("%s not found in archive", tocFilename)
		return &iter
	}

	rc, err := f.Open()
	if err != nil {
		iter.err = fmt.Errorf("failed to open archived %s: %w", tocFilename, err)
		return &iter
	}
	defer rc.Close()

	var nom nomenclature
	if err = json.NewDecoder(rc).Decode(&nom); err != nil {
		iter.err = fmt.Errorf("failed to unmarshal archived %s: %w", tocFilename, err)
		return &iter
	}

	timeline := make([]*igshelf.Media, 0, len(nom.Videos)+len(nom.Photos))
	for _, raw := range nom.Photos {
		m := igshelf.Media{
			Caption:  raw.Caption,
			Type:     igshelf.MediaTypeImage,
			Location: raw.Path,
			TakenAt:  raw.TakenAt,
		}
		// Assign file names which should be used after extracting the files from archive.
		// Year/month prefix helps to explore files.
		_, fname := filepath.Split(raw.Path)
		m.ID = fname[:len(fname)-len(filepath.Ext(fname))]
		m.Filename = m.TakenAt.Format("200601_") + fname
		timeline = append(timeline, &m)
	}
	for _, raw := range nom.Videos {
		m := igshelf.Media{
			Caption:  raw.Caption,
			Type:     igshelf.MediaTypeVideo,
			Location: raw.Path,
			TakenAt:  raw.TakenAt,
		}
		_, fname := filepath.Split(raw.Path)
		m.ID = fname[:len(fname)-len(filepath.Ext(fname))]
		m.Filename = m.TakenAt.Format("200601_") + fname
		timeline = append(timeline, &m)
	}

	// Sort all the media by date to allow grouping by caption.
	// This helps to create albums in MediaIter.
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].TakenAt.After(timeline[j].TakenAt)
	})

	iter.timeline = timeline
	return &iter
}

// MediaIter is an iterator for media timeline.
type MediaIter struct {
	err error
	ctx context.Context
	// cursor is a current cursor position in the timeline slice.
	cursor int
	// current is a current media returned by this iterator.
	current *igshelf.Media
	// timeline is a flat Instagram timeline that must be ordered by date.
	// Otherwise grouping in albums won't work.
	timeline []*igshelf.Media
}

// Next prepares the next media for reading with the Media method.
// It returns true on success, or false if there is no next result or an error
// happened while preparing it. Err should be consulted to distinguish between the two cases.
// Every call to Media, even the first one, must be preceded by a call to Next.
func (it *MediaIter) Next() bool {
	if it.err != nil || it.ctx.Err() != nil {
		return false
	}

	// The cursor is shifted accordingly to skip the children of the current media if it has any.
	m := it.Media()
	if m != nil {
		if len(m.Children) > 0 {
			it.cursor += len(m.Children)
		} else {
			it.cursor++
		}
	}

	if it.cursor >= len(it.timeline) {
		return false
	}

	// When the next few media belong to the same album (dates and captions match), a carousel album is created.
	// Note, ID of this album media is given a suffix to make sure all media IDs are unique.
	m = it.timeline[it.cursor]
	offset := 0
	for i := it.cursor + 1; i < len(it.timeline); i++ {
		if !m.TakenAt.Equal(it.timeline[i].TakenAt) || m.Caption != it.timeline[i].Caption {
			break
		}
		offset = i
	}

	if offset > it.cursor {
		it.current = &igshelf.Media{
			ID:       m.ID + "album",
			Type:     igshelf.MediaTypeAlbum,
			Caption:  m.Caption,
			TakenAt:  m.TakenAt,
			Children: it.timeline[it.cursor : offset+1],
		}
	} else {
		it.current = m
	}
	return true
}

// Media returns the media which the iterator is currently pointing to.
func (it *MediaIter) Media() *igshelf.Media {
	return it.current
}

// Err returns the error, if any, that was encountered during iteration.
func (it *MediaIter) Err() error { return it.err }
