package instagram

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/marselester/igshelf"
)

func TestMediaList_errors(t *testing.T) {
	tt := map[string]struct {
		body       string
		statusCode int
		want       string
	}{
		"nonexisting field": {
			body: `{"error": {
	"message": "Tried accessing nonexisting field (media_type) on node type (User)",
	"type": "IGApiException",
	"code": 100,
	"fbtrace_id": "AT_sdfg081234CQ456-YY"
}}`,
			statusCode: http.StatusBadRequest,
			want:       "IGApiException 100: Tried accessing nonexisting field (media_type) on node type (User)",
		},
		"session expired": {
			body: `{"error": {
    "message": "Error validating access token: Session has expired on Wednesday, 07-Oct-20 18:00:00 PDT. The current time is Thursday, 08-Oct-20 18:36:12 PDT.",
    "type": "OAuthException",
    "code": 190,
	"fbtrace_id": "AT_sdfg082345df5676-ZZ"
}}`,
			statusCode: http.StatusBadRequest,
			want:       "OAuthException 190: Error validating access token: Session has expired on Wednesday, 07-Oct-20 18:00:00 PDT. The current time is Thursday, 08-Oct-20 18:36:12 PDT.",
		},
		"json error": {
			body:       "{",
			statusCode: http.StatusOK,
			want:       "unexpected end of JSON input",
		},
		"empty response": {
			body:       "",
			statusCode: http.StatusInternalServerError,
			want:       "unexpected end of JSON input",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tc.body, tc.statusCode)
			}))
			defer srv.Close()

			client := NewClient("IGQVJ...", WithBaseURL(srv.URL))
			s := NewService(client, "me")

			iter := s.List(context.Background())
			iter.Next()
			if diff := cmp.Diff(tc.want, iter.Err().Error()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestMediaList(t *testing.T) {
	filename := filepath.Join("testdata", "media_list.json")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}))
	defer srv.Close()

	client := NewClient("IGQVJ...", WithBaseURL(srv.URL))
	s := NewService(client, "me")

	iter := s.List(context.Background())
	var got *igshelf.Media
	for iter.Next() {
		got = iter.Media()
	}
	if iter.Err() != nil {
		t.Fatal(iter.Err())
	}

	want := &igshelf.Media{
		ID:        "17850307850323541",
		Caption:   "Still jumping",
		Type:      "CAROUSEL_ALBUM",
		Location:  "https://scontent.cdninstagram.com/v/t51.29350-15/...",
		Filename:  "202010_17850307850323541.jpg",
		Permalink: "https://www.instagram.com/p/CGDFCNqHJv1/",
		TakenAt:   time.Date(2020, time.October, 7, 15, 55, 33, 0, time.UTC),
		Children: []*igshelf.Media{
			{
				ID:                "17850885734317674",
				Type:              "VIDEO",
				Location:          "https://video.cdninstagram.com/v/t50.2886-16/1...",
				ThumbnailLocation: "https://scontent.cdninstagram.com/v/t51.29350-15/1...",
				Filename:          "202010_17850885734317674.mp4",
				ThumbnailFilename: "202010_17850885734317674_cover.jpg",
			},
			{
				ID:                "17863188140095492",
				Type:              "VIDEO",
				Location:          "https://video.cdninstagram.com/v/t50.2886-16/2...",
				ThumbnailLocation: "https://scontent.cdninstagram.com/v/t51.29350-15/2...",
				Filename:          "202010_17863188140095492.mp4",
				ThumbnailFilename: "202010_17863188140095492_cover.jpg",
			},
			{
				ID:                "17871183211965376",
				Type:              "VIDEO",
				Location:          "https://video.cdninstagram.com/v/t50.2886-16/3...",
				ThumbnailLocation: "https://scontent.cdninstagram.com/v/t51.29350-15/3...",
				Filename:          "202010_17871183211965376.mp4",
				ThumbnailFilename: "202010_17871183211965376_cover.jpg",
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf(diff)
	}
}
