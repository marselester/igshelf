# Instagram shelf 📷

As the author of photos/videos I've uploaded to Instagram, I want to have a copy of my content.
The easiest way is to download account data using https://www.instagram.com/download/request/ (zip archive)
and render my Instagram timeline as an html page.
For example, the following command will copy media files from a zip archive to `./content/` directory
and create `timeline.html` (a gallery), `timeline.json` (metadata).

```sh
$ go build ./cmd/igshelf
$ ./igshelf -src=~/Downloads/marselester_20201007.zip
```

I can tweak `template/timeline.tpl` template and render a gallery from existing `timeline.json`.

```sh
$ ./igshelf
```

It may take up to 48 hours to get a link to an Instagram account data.
If I don't want to wait, there is an option to download the content using
[Instagram API](https://developers.facebook.com/docs/instagram-basic-display-api/getting-started).

```sh
$ read -p "Enter access token: " -s IGSHELF_TOKEN
$ export IGSHELF_TOKEN
$ ./igshelf -src=api
```
