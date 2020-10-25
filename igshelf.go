// Package igshelf defines a domain model of a local gallery based on user's Instagram content
// obtained using Instagram Basic Display API or from a zip archive.
package igshelf

import (
	"context"
	"time"
)

const (
	// MediaTypeImage indicates the media is an image.
	MediaTypeImage = "IMAGE"
	// MediaTypeVideo indicates the media is a video.
	MediaTypeVideo = "VIDEO"
	// MediaTypeAlbum indicates the media is a carousel album (images and videos published together).
	MediaTypeAlbum = "CAROUSEL_ALBUM"
)

// Media represents an image, video, or album.
// Album is a media that has a list of images and videos published together.
type Media struct {
	// ID is an identifier of this media in unspecified format, treat it as an arbitrary string.
	// For example, it would look like 17850307850323541 if a media was retrieved from Instagram API.
	// In Instagram zip archive it is 8c996aa535f0f7a322d4dbaef9cfd266.
	// Note, archive doesn't have a notion of album, so IDs are generated locally.
	ID string
	// Type is the media's type. It can be IMAGE, VIDEO, or CAROUSEL_ALBUM.
	Type string
	// Caption is the media's caption text, e.g., "Still jumping".
	Caption string
	// Location is where the file was copied from (either URL or a path in zip archive).
	// For example, https://scontent.cdninstagram.com/v/t51.29350-15/121...01_n.jpg
	// or videos/202010/8c996aa535f0f7a322d4dbaef9cfd266.mp4.
	Location string
	// ThumbnailLocation is a location of video cover (thumbnail image).
	// Note, zip archive doesn't contain thumbnails.
	ThumbnailLocation string
	// Filename is a name of the media file given locally (17841752650018177.mp4) after copying it from original location.
	Filename string
	// ThumbnailFilename is a name of video's thumbnail image given locally, e.g., 17841752650018177_cover.jpg.
	ThumbnailFilename string
	// Permalink is the media's permanent URL, e.g., https://www.instagram.com/p/CGDFCNqHJv1/.
	// Note, Instagram archive doesn't provide permalinks.
	Permalink string
	// TakenAt is the media's publish date.
	TakenAt time.Time
	// Children is a list of media that belong to this media album (CAROUSEL_ALBUM media type).
	// Note, archive doesn't have a notion of album, so igshelf groups photos/videos in albums by their date and caption.
	Children []*Media
}

// MediaService provides access to Instagram timeline so one can get a copy of own content.
type MediaService interface {
	// List returns an iterator that yields media in reverse chronological order (newest first).
	List(ctx context.Context) (iter MediaIter)
	// Download copies the media file and video thumbnail from their location.
	Download(ctx context.Context, m *Media) (content, thumbnail []byte, err error)
}

// MediaIter is an iterator which yields media in reverse chronological order (newest first).
type MediaIter interface {
	// Next prepares the next media for reading with the Media method.
	// It returns true on success, or false if there is no next result or an error
	// happened while preparing it. Err should be consulted to distinguish between the two cases.
	// Every call to Media, even the first one, must be preceded by a call to Next.
	Next() bool
	// Media returns the media which the iterator is currently pointing to.
	Media() *Media
	// Err returns the error, if any, that was encountered during iteration.
	Err() error
}

// MediaRepository is used to store Instagram timeline with assumption that
// a user doesn't have a lot of content (timeline is loaded and stored all at once).
type MediaRepository interface {
	// List returns all the media description as it was stored.
	List() (timeline []*Media, err error)
	// Store persists the media timeline, e.g., as a JSON file.
	Store(timeline []*Media) error
}
