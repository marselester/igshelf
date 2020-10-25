<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
* {
  box-sizing: border-box;
}

body {
  background-color: #f1f1f1;
  padding: 20px;
  font-family: Arial;
}

.row {
  margin: 8px -16px;
}

.row,
.row > .column {
  padding: 8px;
}

.column {
  float: left;
  width: 33%;
}

.row:after {
  content: "";
  display: table;
  clear: both;
}

.content {
  background-color: white;
  padding: 10px;
}

@media screen and (max-width: 1200px) {
  .column {
    width: 50%;
  }
}

@media screen and (max-width: 750px) {
  .column {
    width: 100%;
  }
}
</style>
</head>
<body>
<div class="row">
{{range $post := .Posts}}
  {{- range $i, $_ := .Children}}
  <div class="column">
    <div class="content">
      {{if eq .Type "VIDEO"}}
      <video controls style="max-width:100%; height:auto;" {{if .ThumbnailFilename}}poster="content/{{.ThumbnailFilename}}"{{end}}>
        <source src="content/{{.Filename}}" type="video/mp4">
      </video>
      {{else}}
      <a href="content/{{.Filename}}">
        <img src="content/{{.Filename}}" style="max-width:100%; height:auto;">
      </a>
      {{end}}
      {{if not $i}}
      <p>{{$post.Caption}} <small><a href="content/{{.Filename}}">{{$post.TakenAt.Format "02 Jan 2006"}}</a></small></p>
      {{else}}
      <p><small><a href="content/{{.Filename}}">{{$post.TakenAt.Format "02 Jan 2006"}}</a></small></p>
      {{end}}
    </div>
  </div>
  {{else}}
  <div class="column">
    <div class="content">
      {{if eq .Type "VIDEO"}}
      <video controls style="max-width:100%; height:auto;" {{if .ThumbnailFilename}}poster="content/{{.ThumbnailFilename}}"{{end}}>
        <source src="content/{{.Filename}}" type="video/mp4">
      </video>
      {{else}}
      <a href="content/{{.Filename}}">
        <img src="content/{{.Filename}}" style="max-width:100%; height:auto;">
      </a>
      {{end}}
      <p>{{ .Caption}} <small><a href="content/{{.Filename}}">{{.TakenAt.Format "02 Jan 2006"}}</a></small></p>
    </div>
  </div>
  {{end -}}
{{end}}
</div>
</body>
</html>
