// Package mock provides mocks for igshelf domain to facilitate testing.
package mock

import (
	"context"

	"github.com/marselester/igshelf"
)

// MediaService is a mock that implements igshelf.MediaService interface.
type MediaService struct {
	ListFn     func() (iter igshelf.MediaIter)
	DownloadFn func(m *igshelf.Media) (content, thumbnail []byte, err error)
}

// List calls ListFn to inspect the mock.
func (s *MediaService) List(ctx context.Context) (iter igshelf.MediaIter) {
	if s.ListFn == nil {
		return nil
	}
	return s.ListFn()
}

// Download calls DownloadFn to inspect the mock.
func (s *MediaService) Download(ctx context.Context, m *igshelf.Media) (content, thumbnail []byte, err error) {
	if s.DownloadFn == nil {
		return nil, nil, nil
	}
	return s.DownloadFn(m)
}

// MediaIter is a mock that implements igshelf.MediaIter interface.
type MediaIter struct {
	// Batch of media the iterator will work with by default.
	Batch []*igshelf.Media
	err   error
	// cursor is a current cursor position in Batch.
	cursor int
	// current is a current media returned by this iterator.
	current *igshelf.Media

	NextFn  func() bool
	MediaFn func() *igshelf.Media
	ErrFn   func() error
}

// Next calls NextFn to inspect the mock if the func was configured.
// Otherwise it provides an iterator over the Batch implementation.
func (it *MediaIter) Next() bool {
	if it.NextFn != nil {
		return it.NextFn()
	}

	if it.Media() != nil {
		it.cursor++
	}

	if it.cursor >= len(it.Batch) {
		return false
	}

	it.current = it.Batch[it.cursor]
	return true
}

// Media calls MediaFn to inspect the mock if the func was configured.
// Otherwise it provides an iterator over the Batch implementation.
func (it *MediaIter) Media() *igshelf.Media {
	if it.MediaFn != nil {
		it.MediaFn()
	}

	return it.current
}

// Err calls ErrFn to inspect the mock if the func was configured.
// Otherwise it provides an iterator over the Batch implementation.
func (it *MediaIter) Err() error {
	if it.ErrFn != nil {
		return it.ErrFn()
	}

	return it.err
}

// MediaRepository is a mock that implements igshelf.MediaRepository interface.
type MediaRepository struct {
	ListFn  func() (timeline []*igshelf.Media, err error)
	StoreFn func(timeline []*igshelf.Media) error
}

// List calls ListFn to inspect the mock.
func (r *MediaRepository) List() (timeline []*igshelf.Media, err error) {
	if r.ListFn == nil {
		return nil, nil
	}
	return r.ListFn()
}

// Store calls StoreFn to inspect the mock.
func (r *MediaRepository) Store(timeline []*igshelf.Media) error {
	if r.StoreFn == nil {
		return nil
	}
	return r.StoreFn(timeline)
}
