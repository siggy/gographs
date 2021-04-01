[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/siggy/gographs)

# gographs

[gographs](https://gographs.io) renders dependency graphs for Go packages.

![gographs dependency example](https://gographs.io/repo/github.com/siggy/gographs.svg "gographs dependencies")

## Badge Markdown

```md
[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/GO_REPO)
```

Example:
```md
[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/siggy/gographs)
```

Courtesy of [shields.io]([https://shields.io/])

## HTTP Endpoints

### [gographs.io](https://gographs.io)
Defaults to rendering this Go repo.

`text/html; charset=utf-8`

### [gographs.io/repo/GO_REPO?cluster=false|true](https://gographs.io/repo/github.com/siggy/gographs)
Permalink to view a Go repo SVG.

`text/html; charset=utf-8`

To refresh: use `POST`.

### [gographs.io/graph/GO_REPO.svg?cluster=false|true](https://gographs.io/graph/github.com/siggy/gographs.svg)
Go repo SVG direct link.

`image/svg+xml; charset=utf-8`

To refresh: use `POST`.

### [gographs.io/graph/GO_REPO.dot?cluster=false|true](https://gographs.io/graph/github.com/siggy/gographs.dot)
Go repo GraphViz DOT direct link.

`text/plain; charset=utf-8`

To refresh: use `POST`.

### [gographs.io/svg?url=SVG_URL](https://gographs.io/svg?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg)
Permalink to view an arbitrary SVG URL.

`text/html; charset=utf-8`

## Local dev

### First-time setup

```bash
go install github.com/loov/goda@v0.2.2
brew install dot # or equivalent
brew install redis # or equivalent
redis-server /usr/local/etc/redis.conf
```

### Boot server

```bash
go run main.go --log-level debug
```

Browse to http://localhost:8888

## Lint check

```bash
bin/lint
```

## Credits

This tool is built using many open source packages, but two in particular
deserve calling out:

- [goda](https://github.com/loov/goda)
- [SVGPan](https://github.com/ariutta/svg-pan-zoom)

The [`pkg/repo`](./pkg/repo) is based on:

- [Go Report Card](https://github.com/gojp/goreportcard)
