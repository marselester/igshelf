package downloader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/marselester/igshelf"
	"github.com/marselester/igshelf/internal/mock"
)

func TestTimelineIsStored(t *testing.T) {
	want := []*igshelf.Media{{
		ID:                "17863188140095492",
		Type:              "VIDEO",
		Location:          "https://video.cdninstagram.com/v/t50.2886-16/2...",
		ThumbnailLocation: "https://scontent.cdninstagram.com/v/t51.29350-15/2...",
		Filename:          "17863188140095492.mp4",
		ThumbnailFilename: "17863188140095492_cover.jpg",
	}}

	ig := mock.MediaService{
		ListFn: func() igshelf.MediaIter {
			return &mock.MediaIter{Batch: want}
		},
		DownloadFn: func(m *igshelf.Media) ([]byte, []byte, error) {
			return nil, nil, fmt.Errorf("boom")
		},
	}
	db := mock.MediaRepository{StoreFn: func(got []*igshelf.Media) error {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(diff)
		}
		return nil
	}}
	s := NewService(&ig, &db)

	err := s.Download(context.Background(), "testdata")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownload(t *testing.T) {
	timeline := []*igshelf.Media{{
		ID:                "17863188140095492",
		Type:              "VIDEO",
		Location:          "https://video.cdninstagram.com/v/t50.2886-16/2...",
		ThumbnailLocation: "https://scontent.cdninstagram.com/v/t51.29350-15/2...",
		Filename:          "17863188140095492.mp4",
		ThumbnailFilename: "17863188140095492_cover.jpg",
	}}
	var (
		wantFile  = []byte("content")
		wantThumb = []byte("thumbnail")
	)
	t.Cleanup(func() {
		if err := os.Remove("testdata/17863188140095492.mp4"); err != nil {
			t.Errorf("failed to delete a video: %w", err)
		}
		if err := os.Remove("testdata/17863188140095492_cover.jpg"); err != nil {
			t.Errorf("failed to delete a video cover: %w", err)
		}
	})

	ig := mock.MediaService{
		ListFn: func() igshelf.MediaIter {
			return &mock.MediaIter{Batch: timeline}
		},
		DownloadFn: func(m *igshelf.Media) ([]byte, []byte, error) {
			return wantFile, wantThumb, nil
		},
	}
	db := mock.MediaRepository{}
	s := NewService(&ig, &db)

	err := s.Download(context.Background(), "testdata")
	if err != nil {
		t.Fatal(err)
	}

	got, err := ioutil.ReadFile("testdata/17863188140095492.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantFile, got); diff != "" {
		t.Errorf(diff)
	}

	got, err = ioutil.ReadFile("testdata/17863188140095492_cover.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantThumb, got); diff != "" {
		t.Errorf(diff)
	}
}
