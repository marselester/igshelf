// Program igshelf creates a local gallery (timeline.html) of user's Instagram content.
// It can work with Instagram zip archive and Instagram Basic Display API.
// Note, the program doesn't stop if one of the files was not copied due to an error.
// For example, media.json might list a photo which actually wasn't included into a zip archive.
// There were four missing photos/videos when author received his zip archive.
package main

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/peterbourgon/ff/v3"

	"github.com/marselester/igshelf"
	"github.com/marselester/igshelf/internal/archive"
	"github.com/marselester/igshelf/internal/downloader"
	"github.com/marselester/igshelf/internal/instagram"
	"github.com/marselester/igshelf/internal/jsonfile"
)

func main() {
	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// By default an exit code is set to indicate a failure
	// since there are more failure scenarios to begin with.
	exitCode := 1
	defer func() { os.Exit(exitCode) }()

	fs := flag.NewFlagSet("igshelf", flag.ExitOnError)
	var (
		source      = fs.String("src", "", `source of the Instagram timeline ("api" or path to a zip archive)`)
		destination = fs.String("dst", "", "path to a directory where timeline is stored")
		workerCount = fs.Int("worker", 10, "number of workers that copy media files")
		token       = fs.String("token", "", "Instagram API access token")
		user        = fs.String("user", "me", "user whose timeline should be downloaded")
		_           = fs.String("config", "", "config file")
	)
	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("IGSHELF"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		logger.Log("msg", "failed to parse flags", "err", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sig:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	var (
		timelineJSONpath = filepath.Join(*destination, "timeline.json")
		timelineHTMLpath = filepath.Join(*destination, "timeline.html")
		contentDirPath   = filepath.Join(*destination, "content")
		templatePath     = filepath.Join("template", "timeline.tpl")
	)
	// Create a directory to store media files.
	_, err = os.Stat(contentDirPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(contentDirPath, 0700); err != nil {
			logger.Log("msg", "failed to create content dir", "path", contentDirPath, "err", err)
			return
		}
	}

	db := jsonfile.NewMediaRepository(timelineJSONpath)

	// Prepare a media service in case a user decides to download the timeline
	// from API or a zip archive.
	var ig igshelf.MediaService
	switch {
	case *source == "api":
		ig = instagram.NewService(
			instagram.NewClient(*token),
			*user,
		)
	case strings.HasSuffix(*source, ".zip"):
		arch, err := archive.NewService(*source)
		if err != nil {
			logger.Log("msg", "failed to open Instagram archive", "err", err)
			return
		}
		defer arch.Close()
		ig = arch
	}

	var timeline []*igshelf.Media
	// Fetch user's timeline and store timeline.json in the destination directory
	// along with downloaded media files (photos, videos).
	if ig != nil {
		d := downloader.NewService(ig, db,
			downloader.WithMaxWorkers(*workerCount),
			downloader.WithLogger(logger),
		)
		err = d.Download(ctx, contentDirPath)
		if err != nil {
			logger.Log("msg", "failed to download the timeline", "err", err)
			return
		}
	}

	// Read existing timeline.json from the destination directory.
	if timeline, err = db.List(); err != nil {
		logger.Log("msg", "failed to read the local timeline", "err", err)
		return
	}

	// Render the timeline as html page.
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Log("msg", "failed to parse the template", "path", templatePath, "err", err)
		return
	}

	data := struct {
		Posts []*igshelf.Media
	}{timeline}
	buf := bytes.Buffer{}
	if err = t.Execute(&buf, data); err != nil {
		logger.Log("msg", "failed to render the timeline", "err", err)
		return
	}
	if err = ioutil.WriteFile(timelineHTMLpath, buf.Bytes(), 0600); err != nil {
		logger.Log("msg", "failed to write timeline.html on disk", "err", err)
		return
	}

	// The program terminates successfully.
	exitCode = 0
}
