package downloader

import "github.com/go-kit/kit/log"

// ConfigOption configures the downloader service.
type ConfigOption func(*Service)

// WithMaxWorkers sets a max limit of workers to spawn when downloading photos/videos.
func WithMaxWorkers(n int) ConfigOption {
	return func(s *Service) {
		s.maxWorkers = n
	}
}

// WithLogger configures a logger to debug media files downloading.
func WithLogger(l log.Logger) ConfigOption {
	return func(r *Service) {
		r.logger = l
	}
}
