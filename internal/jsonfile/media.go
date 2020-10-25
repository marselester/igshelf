// Package jsonfile provides JSON file based repository implementation to store Instagram timeline.
package jsonfile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/marselester/igshelf"
)

// MediaRepository stores Instagram timeline in a JSON file.
type MediaRepository struct {
	filename string
}

// NewMediaRepository creates new MediaRepository.
func NewMediaRepository(filename string) *MediaRepository {
	return &MediaRepository{filename: filename}
}

// List returns all the media description as it was stored.
func (r *MediaRepository) List() ([]*igshelf.Media, error) {
	b, err := ioutil.ReadFile(r.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read timeline from disk %s: %w", r.filename, err)
	}

	var timeline []*igshelf.Media
	if err = json.Unmarshal(b, &timeline); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timeline %s: %w", r.filename, err)
	}
	return timeline, nil
}

// Store persists the media timeline on disk.
// The file is always overwritten.
func (r *MediaRepository) Store(timeline []*igshelf.Media) error {
	b, err := json.Marshal(&timeline)
	if err != nil {
		return fmt.Errorf("failed to marshal timeline %s: %w", r.filename, err)
	}

	if err = ioutil.WriteFile(r.filename, b, 0600); err != nil {
		return fmt.Errorf("failed to write timeline on disk %s: %w", r.filename, err)
	}

	return nil
}
