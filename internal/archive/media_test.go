package archive

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/marselester/igshelf"
)

func TestMediaList(t *testing.T) {
	filename := filepath.Join("testdata", "marselester_20201007.zip")
	arch, err := NewService(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer arch.Close()

	iter := arch.List(context.Background())
	var got []*igshelf.Media
	for iter.Next() {
		got = append(got, iter.Media())
	}
	if iter.Err() != nil {
		t.Fatal(iter.Err())
	}

	want := []*igshelf.Media{
		{
			ID:       "8c996aa535f0f7a322d4dbaef9cfd266",
			Caption:  "Still jumping",
			Type:     "VIDEO",
			Location: "videos/202010/8c996aa535f0f7a322d4dbaef9cfd266.mp4",
			Filename: "202010_8c996aa535f0f7a322d4dbaef9cfd266.mp4",
			TakenAt:  time.Date(2020, time.October, 7, 15, 55, 33, 0, time.UTC),
		},
		{
			ID:       "d8612ffa060b392077322ccf2e953f35",
			Caption:  "Starting another two-wheeled hobby.\n\nЯ буду долго гнать велосипед.",
			Type:     "IMAGE",
			Location: "photos/202006/d8612ffa060b392077322ccf2e953f35.jpg",
			Filename: "202006_d8612ffa060b392077322ccf2e953f35.jpg",
			TakenAt:  time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(diff)
	}
}

func TestMediaIter(t *testing.T) {
	tt := map[string]struct {
		timeline []*igshelf.Media
		want     []*igshelf.Media
	}{
		"nil": {
			timeline: nil,
			want:     nil,
		},
		"blank": {
			timeline: make([]*igshelf.Media, 0),
			want:     nil,
		},
		"image": {
			timeline: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "IMAGE",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
			},
			want: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "IMAGE",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
			},
		},
		"album": {
			timeline: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "IMAGE",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "2",
					Type:    "VIDEO",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
			},
			want: []*igshelf.Media{
				{
					ID:      "1album",
					Type:    "CAROUSEL_ALBUM",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
					Children: []*igshelf.Media{
						{
							ID:      "1",
							Type:    "IMAGE",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
						{
							ID:      "2",
							Type:    "VIDEO",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
					},
				},
			},
		},
		"video and album": {
			timeline: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "VIDEO",
					Caption: "cats",
					TakenAt: time.Date(2020, time.October, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "2",
					Type:    "IMAGE",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "3",
					Type:    "VIDEO",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
			},
			want: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "VIDEO",
					Caption: "cats",
					TakenAt: time.Date(2020, time.October, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "2album",
					Type:    "CAROUSEL_ALBUM",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
					Children: []*igshelf.Media{
						{
							ID:      "2",
							Type:    "IMAGE",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
						{
							ID:      "3",
							Type:    "VIDEO",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
					},
				},
			},
		},
		"album and image": {
			timeline: []*igshelf.Media{
				{
					ID:      "1",
					Type:    "IMAGE",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "2",
					Type:    "VIDEO",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
				},
				{
					ID:      "3",
					Type:    "IMAGE",
					Caption: "cats",
					TakenAt: time.Date(2020, time.March, 21, 1, 12, 14, 0, time.UTC),
				},
			},
			want: []*igshelf.Media{
				{
					ID:      "1album",
					Type:    "CAROUSEL_ALBUM",
					Caption: "still jumping",
					TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
					Children: []*igshelf.Media{
						{
							ID:      "1",
							Type:    "IMAGE",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
						{
							ID:      "2",
							Type:    "VIDEO",
							Caption: "still jumping",
							TakenAt: time.Date(2020, time.June, 21, 1, 12, 14, 0, time.UTC),
						},
					},
				},
				{
					ID:      "3",
					Type:    "IMAGE",
					Caption: "cats",
					TakenAt: time.Date(2020, time.March, 21, 1, 12, 14, 0, time.UTC),
				},
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			iter := MediaIter{
				ctx:      context.Background(),
				timeline: tc.timeline,
			}

			var got []*igshelf.Media
			for iter.Next() {
				got = append(got, iter.Media())
			}
			if iter.Err() != nil {
				t.Fatal(iter.Err())
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
