# GoGraphs

[GoGraphs](https://gographs.io) renders dependency graphs for Go packages.

![GoGraphs dependency example](https://gographs.io/repo/github.com/siggy/gographs.svg "GoGraphs Dependencies")

## HTTP Endpoints

### [gographs.io](https://gographs.io)
Defaults to rendering this Go repo.

`text/html; charset=utf-8`

### [gographs.io/repo/GO_REPO?cluster=false|true](https://gographs.io/repo/github.com/siggy/gographs)
Permalink to view a Go repo SVG.

`text/html; charset=utf-8`

### [gographs.io/graph/GO_REPO.svg?cluster=false|true](https://gographs.io/graph/github.com/siggy/gographs.svg)
Go repo SVG direct link.

`image/svg+xml; charset=utf-8`

### [gographs.io/graph/GO_REPO.dot?cluster=false|true](https://gographs.io/graph/github.com/siggy/gographs.dot)
Go repo GraphViz DOT direct link.

`text/plain; charset=utf-8`

### [gographs.io/svg?url=SVG_URL](https://gographs.io/svg?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg)
Permalink to view an arbitrary SVG URL.

`text/html; charset=utf-8`

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

- warning on spinner requests can take up to 1 minute
- handle form input that includes the `http[s]://`
- show "Last refresh" time
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
- fix new svg loading on firefox
- prevent normal dragging on mobile
- fix thumbnail resize/refresh jitter
- fix irc logo thumbnail border aspect ratio
- show progress during repo compile
- generic svg and dot support, paste and url
- consider capturing all mouse scrolling over every element
- on drag mouseup, don't open godoc link
