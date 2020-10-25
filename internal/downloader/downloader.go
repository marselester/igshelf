// Package downloader provides a service to copy Instagram timeline (photos/videos).
package downloader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"golang.org/x/sync/errgroup"

	"github.com/marselester/igshelf"
)

const (
	// defaultMaxWorkers is a max number of workers to spawn when downloading media files.
	defaultMaxWorkers = 10
)

// Service is a service that copies Instagram timeline using media service
// and persists it with a media repository.
type Service struct {
	ig     igshelf.MediaService
	db     igshelf.MediaRepository
	logger log.Logger

	maxWorkers int
	// sem is a semaphore that limits count of workers that copy media files.
	// Acquire this semaphore by sending a token, and release it by discarding a token.
	sem chan token
}
type token struct{}

// NewService creates a service to copy Instagram timeline.
func NewService(ig igshelf.MediaService, db igshelf.MediaRepository, options ...ConfigOption) *Service {
	s := Service{
		ig:     ig,
		db:     db,
		logger: log.NewNopLogger(),

		maxWorkers: defaultMaxWorkers,
	}
	for _, opt := range options {
		opt(&s)
	}
	s.sem = make(chan token, s.maxWorkers)
	return &s
}

// Download fetches the timeline using Instagram media service (e.g., zip archive or Instagram API)
// and stores it in a JSON file.
// After that it copies media files concurrently.
// It doesn't stop if one of the files was not copied due to an error.
// For example, media.json might list a file which actually wasn't included into the archive.
func (s *Service) Download(ctx context.Context, contentDirPath string) error {
	var timeline []*igshelf.Media
	iter := s.ig.List(ctx)
	for iter.Next() {
		timeline = append(timeline, iter.Media())
	}
	if iter.Err() != nil {
		return fmt.Errorf("failed to fetch the timeline: %w", iter.Err())
	}

	if err := s.db.Store(timeline); err != nil {
		return fmt.Errorf("failed to store the timeline: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	mediac := make(chan *igshelf.Media, s.maxWorkers)

	// Line up all the media (including children) for downloading.
	g.Go(func() error {
		defer close(mediac)

		for _, m := range timeline {
			select {
			case mediac <- m:
			case <-ctx.Done():
				return ctx.Err()
			}

			for _, m = range m.Children {
				select {
				case mediac <- m:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		return nil
	})

	for m := range mediac {
		// There will be only one idle worker instead of many,
		// because goroutines shouldn't sit around doing nothing.
		s.sem <- token{}

		m := m
		g.Go(func() error {
			defer func() { <-s.sem }()

			// Zip archive doesn't contain albums.
			if m.Filename == "" {
				return nil
			}

			contentPath := filepath.Join(contentDirPath, m.Filename)
			// Skip downloading if the media file already exists.
			_, err := os.Stat(contentPath)
			if os.IsExist(err) {
				return nil
			}

			content, thumbnail, err := s.ig.Download(ctx, m)
			if err != nil {
				s.logger.Log("msg", "failed to download media content", "media", m, "err", err)
				return nil
			}

			if err = ioutil.WriteFile(contentPath, content, 0600); err != nil {
				return fmt.Errorf("failed to store media content %s: %w", m.ID, err)
			}

			if thumbnail != nil {
				thumbnailPath := filepath.Join(contentDirPath, m.ThumbnailFilename)
				if err = ioutil.WriteFile(thumbnailPath, thumbnail, 0600); err != nil {
					return fmt.Errorf("failed to store media thumbnail %s: %w", m.ID, err)
				}
			}

			return nil
		})
	}

	return g.Wait()
}
