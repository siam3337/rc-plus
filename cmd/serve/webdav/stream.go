package webdav

// Put this file next to webdav.go in cmd/serve/webdav/ (package name must match
// the other files there; change "webdav" above if your checkout differs).
//
//   GET /path/ABC.mp4            -> normal file / WebDAV behaviour
//   GET /path/ABC.mp4?stream     -> standalone Plyr player page
//   GET /path/ABC.mp4?=stream    -> same (any query containing "stream")
//
// Call it from (w *WebDAV) ServeHTTP in webdav.go, see the comment there.

import (
	"html"
	"net/http"
	"path"
	"strings"
)

var streamExts = map[string]bool{
	".mp4": true, ".webm": true, ".mkv": true, ".mov": true, ".m4v": true, ".ogv": true,
	".mp3": true, ".m4a": true, ".aac": true, ".flac": true, ".wav": true, ".ogg": true, ".opus": true,
}

var audioExts = map[string]bool{
	".mp3": true, ".m4a": true, ".aac": true, ".flac": true, ".wav": true, ".ogg": true, ".opus": true,
}

const streamPage = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{TITLE}}</title>
<link rel="stylesheet" href="https://cdn.plyr.io/3.7.8/plyr.css">
<style>
html, body { margin: 0; height: 100%; background: #000; font-family: system-ui, sans-serif; }
body { display: flex; flex-direction: column; }
.bar { display: flex; gap: 12px; align-items: center; padding: 10px 14px; background: #14151d; color: #f1f2f7; font-size: 14px; }
.bar .t { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 600; }
.bar a { color: #f1f2f7; background: #262a38; padding: 5px 13px; border-radius: 999px; text-decoration: none; font-size: 13px; }
.bar a:hover { background: #6c5ce7; }
.wrap { flex: 1; min-height: 0; display: flex; align-items: center; justify-content: center; --plyr-color-main: #6c5ce7; }
.wrap .plyr { width: 100%; max-height: 100%; }
</style>
</head>
<body>
<div class="bar"><span class="t">{{TITLE}}</span><a href="{{SRC}}" download>Download</a></div>
<div class="wrap"><{{TAG}} id="p" playsinline controls preload="metadata" src="{{SRC}}"></{{TAG}}></div>
<script src="https://cdn.plyr.io/3.7.8/plyr.polyfilled.js"></script>
<script>
var p = new Plyr('#p', {
  controls: ['play-large','play','progress','current-time','duration','mute','volume','settings','pip','airplay','fullscreen'],
  settings: ['speed'], seekTime: 10
});
</script>
</body>
</html>`

// serveStreamPage writes the Plyr page and returns true when the request is a
// GET for a media file whose query contains "stream". Otherwise it writes
// nothing and returns false. baseURL is the --baseurl value ("/" when unset).
func serveStreamPage(rw http.ResponseWriter, r *http.Request, baseURL string) bool {
	if r.Method != http.MethodGet || !strings.Contains(r.URL.RawQuery, "stream") {
		return false
	}
	ext := strings.ToLower(path.Ext(r.URL.Path))
	if !streamExts[ext] {
		return false
	}
	tag := "video"
	if audioExts[ext] {
		tag = "audio"
	}
	src := strings.TrimRight(baseURL, "/") + r.URL.EscapedPath()
	page := strings.NewReplacer(
		"{{TITLE}}", html.EscapeString(path.Base(r.URL.Path)),
		"{{SRC}}", html.EscapeString(src),
		"{{TAG}}", tag,
	).Replace(streamPage)
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-store")
	_, _ = rw.Write([]byte(page))
	return true
}
