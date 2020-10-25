// Package instagram provides access to user's timeline via Instagram Basic Display API.
package instagram

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/marselester/igshelf"
)

// MediaService provides access to user's Instagram timeline via Instagram Basic Display API.
type MediaService struct {
	client *Client
	userID string
}

// NewService creates a media service that provides access to user's Instagram timeline via Instagram Basic Display API.
// In most cases userID should be "me", but you can also set explicit ID such as 17843400535183040.
func NewService(client *Client, userID string) *MediaService {
	return &MediaService{
		client: client,
		userID: userID,
	}
}

// media represents an image, video, or album requested from Instagram API.
type media struct {
	// ID is the media's ID, e.g., 17850307850323541.
	ID string `json:"id"`
	// Type is the media's type. It can be IMAGE, VIDEO, or CAROUSEL_ALBUM.
	Type string `json:"media_type"`
	// Caption is the media's caption text, e.g., "Still jumping".
	// It is not returnable for media in albums.
	Caption string `json:"caption"`
	// URL is the media's URL, e.g., https://scontent.cdninstagram.com/v/t51.29350-15/...
	URL string `json:"media_url"`
	// Permalink is the media's permanent URL, e.g., https://www.instagram.com/p/CGDFCNqHJv1/.
	// It will be omitted if the media contains copyrighted material,
	// or has been flagged for a copyright violation.
	Permalink string `json:"permalink"`
	// ThumbnailURL is the media's thumbnail image URL. It is only available on VIDEO media.
	ThumbnailURL string `json:"thumbnail_url"`
	// TakenAt is the media's publish date in ISO 8601 format, e.g., 2019-11-10T12:20:51+0000.
	TakenAt timeISO8601 `json:"timestamp"`
	// Children is a list of media on the album. It's only available on CAROUSEL_ALBUM media.
	Children struct {
		Data []*media
	}
}

// timeISO8601 is used to parse Instagram's timestamp field, e.g., 2019-11-10T12:20:51+0000.
type timeISO8601 time.Time

// UnmarshalJSON decodes ISO 8601 time as time.Time.
func (t *timeISO8601) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	v, err := time.Parse("\"2006-01-02T15:04:05+0000\"", string(data))
	if err != nil {
		return err
	}
	*t = timeISO8601(v)
	return nil
}

// Download copies the media file and its thumbnail (video cover) if it's available.
func (s *MediaService) Download(ctx context.Context, m *igshelf.Media) (content, thumbnail []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, m.Location, nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)
	contResp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download content: %w", err)
	}
	defer contResp.Body.Close()
	if content, err = ioutil.ReadAll(contResp.Body); err != nil {
		return nil, nil, fmt.Errorf("failed to read content: %w", err)
	}

	if m.ThumbnailLocation == "" {
		return content, nil, nil
	}

	req, err = http.NewRequest(http.MethodGet, m.ThumbnailLocation, nil)
	if err != nil {
		return content, nil, err
	}
	req = req.WithContext(ctx)
	thumbResp, err := s.client.httpClient.Do(req)
	if err != nil {
		return content, nil, fmt.Errorf("failed to download thumbnail: %w", err)
	}
	defer thumbResp.Body.Close()
	if thumbnail, err = ioutil.ReadAll(thumbResp.Body); err != nil {
		return content, nil, fmt.Errorf("failed to read thumbnail: %w", err)
	}

	return content, thumbnail, nil
}

// mediaListResp is a list of media as retrieved from the media API endpoint.
type mediaListResp struct {
	Batch  []*media `json:"data"`
	Paging struct {
		Cursors struct {
			After string `json:"after"`
		} `json:"cursors"`
	} `json:"paging"`
}

// List returns an iterator to access the user's timeline.
// It relies on API pagination to fetch batches of media.
// You would have to start over if an error occurs during pagination (server timeout).
func (s *MediaService) List(ctx context.Context) igshelf.MediaIter {
	path := fmt.Sprintf("%s/media", s.userID)
	queryParams := url.Values{}
	queryParams.Set("fields", "id,caption,media_type,media_url,permalink,thumbnail_url,timestamp,children{media_type,media_url,thumbnail_url}")

	return &MediaIter{fetch: func() ([]*igshelf.Media, error) {
		// Stop iterator when the next pagination token is empty.
		if _, ok := queryParams["after"]; ok && queryParams.Get("after") == "" {
			return nil, nil
		}

		req, err := s.client.NewRequest(ctx, http.MethodGet, path, queryParams, nil)
		if err != nil {
			return nil, err
		}

		v := mediaListResp{}
		_, err = s.client.Do(req, &v)
		if err != nil {
			return nil, err
		}

		queryParams.Set("after", v.Paging.Cursors.After)

		mm := make([]*igshelf.Media, len(v.Batch))
		for i, raw := range v.Batch {
			mm[i] = &igshelf.Media{
				ID:                raw.ID,
				Type:              raw.Type,
				Caption:           raw.Caption,
				Location:          raw.URL,
				ThumbnailLocation: raw.ThumbnailURL,
				Permalink:         raw.Permalink,
				TakenAt:           time.Time(raw.TakenAt),
			}
			// Assign file names which should be used when storing photos/videos locally.
			// Year/month prefix helps to explore files.
			fname := mm[i].TakenAt.Format("200601_") + raw.ID
			switch raw.Type {
			case igshelf.MediaTypeImage, igshelf.MediaTypeAlbum:
				mm[i].Filename = fname + ".jpg"
			case igshelf.MediaTypeVideo:
				mm[i].Filename = fname + ".mp4"
				mm[i].ThumbnailFilename = fname + "_cover.jpg"
			}

			if len(raw.Children.Data) > 0 {
				mm[i].Children = make([]*igshelf.Media, len(raw.Children.Data))
				for j, c := range raw.Children.Data {
					mm[i].Children[j] = &igshelf.Media{
						ID:                c.ID,
						Type:              c.Type,
						Location:          c.URL,
						ThumbnailLocation: c.ThumbnailURL,
					}
					fname = mm[i].TakenAt.Format("200601_") + c.ID
					switch c.Type {
					case igshelf.MediaTypeImage, igshelf.MediaTypeAlbum:
						mm[i].Children[j].Filename = fname + ".jpg"
					case igshelf.MediaTypeVideo:
						mm[i].Children[j].Filename = fname + ".mp4"
						mm[i].Children[j].ThumbnailFilename = fname + "_cover.jpg"
					}
				}
			}
		}

		return mm, nil
	}}
}

// MediaIter is an iterator for collection of media.
type MediaIter struct {
	fetch  func() ([]*igshelf.Media, error)
	err    error
	cursor int
	batch  []*igshelf.Media
}

// Media returns the media which the iterator is currently pointing to.
func (mi *MediaIter) Media() *igshelf.Media {
	return mi.batch[mi.cursor]
}

// Next prepares the next media for reading with the Media method.
// It returns true on success, or false if there is no next result or an error
// happened while preparing it. Err should be consulted to distinguish between the two cases.
// Every call to Media, even the first one, must be preceded by a call to Next.
func (mi *MediaIter) Next() bool {
	if mi.err != nil {
		return false
	}

	if mi.cursor >= len(mi.batch)-1 {
		mi.cursor = 0
		mi.batch, mi.err = mi.fetch()
		if mi.err != nil || len(mi.batch) == 0 {
			return false
		}
		return true
	}

	mi.cursor++
	return true
}

// Err returns the error, if any, that was encountered during iteration.
func (mi *MediaIter) Err() error { return mi.err }
