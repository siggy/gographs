# GoGraphs

[GoGraphs](https://gographs.io) renders dependency graphs for Go packages.

![GoGraphs dependency example](https://gographs.io/repo/github.com/siggy/gographs.svg "GoGraphs Dependencies")

## HTTP Endpoints

### GET /

Defaults to rendering this Go package.

- Example: https://gographs.io
- Content-Type: `text/html; charset=utf-8`

### GET /repo/[GO_REPO]?cluster=[false|true]

Permalink to view a Go repo SVG.

- Example: https://gographs.io/repo/github.com/siggy/gographs
- Content-Type: `text/html; charset=utf-8`

### GET /graph/[GITHUB_REPO].svg?cluster=[false|true]

Go repo SVG direct link.

- Example: https://gographs.io/graph/github.com/siggy/gographs.svg
- Content-Type: `image/svg+xml; charset=utf-8`

### GET /graph/[GITHUB_REPO].dot?cluster=[false|true]

Go repo GraphViz DOT direct link.

- Example:  https://gographs.io/graph/github.com/siggy/gographs.dot
- Content-Type: `text/plain; charset=utf-8`

### GET /svg?url=[SVG_URL]

Permalink to view an arbitrary SVG URL.

- Example: https://gographs.io/svg?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg
- Content-Type: `text/html; charset=utf-8`

## Local dev

### First-time setup

```bash
go install github.com/loov/goda@v0.2.1
brew install dot # or equivalent
brew install redis # or equivalent
redis-server /usr/local/etc/redis.conf
```

### Boot server

```bash
go run main.go --addr localhost:8888 --redis-addr localhost:6379 --log-level debug
```

Browse to http://localhost:8888

## Testing

```bash
golint ./...
```

## Credits

This tool is built using many open source packages, but two in particular
deserve calling out:

- [goda](https://github.com/loov/goda)
- [SVGPan](https://github.com/ariutta/svg-pan-zoom)

## TODO

- refresh should be POST-only
- new URL scheme gographs.io/repo/github.com/siggy/gographs?cluster=true&refresh=true
  - already a thing: http://localhost:8888/repo/github.com/linkerd/linkerd2.svg?cluster=true
- change godoc to pkg.dev
- make Clear() work
- make fonts bigger / easier to read
- fix double click to zoom changing center / viewport limits
- fix thumbnail click to move not tracking the click point exactly
- fix http://localhost:8888/?repo=github.com/kubernetes/kubernetes&cluster=true
- deterministic repo location, with cli flag
- use bin/goda
- don't default to master.info
- faster, replace Download
  - golang.org/x/tools/go/vcs or
  - https://github.com/gojp/goreportcard/blob/master/download/download.go
  - OR: use existing proxy, then fall back to vcs/download.go
- don't shell out
- landing page
- prod
  - default to this repo
  - remove or repurpose localhost URLs in readme
  - monitoring, cache hits/misses
- graphs by repo revision
- "refresh now" button, show cache age
- fix new svg loading on firefox
- prevent normal dragging on mobile
- refresh button
- fix thumbnail resize/refresh jitter
- fix irc logo thumbnail border aspect ratio
- show progress during repo compile
- generic svg and dot support, paste and url
- consider capturing all mouse scrolling over every element
- on drag mouseup, don't open godoc link
