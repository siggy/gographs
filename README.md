# GoGraphs

[GoGraphs](https://gographs.io) renders dependency graphs for Go packages:

![gographs dependency example](https://gographs.io/repo/github.com/siggy/gographs.svg?cluster=false "GoGraphs Dependencies")

## Credits

This tool is built using many open source packages, but two in particular
deserve calling out:

- [goda](https://github.com/loov/goda)
- [SVGPan](https://github.com/ariutta/svg-pan-zoom)

## Local dev

### First-time setup

```bash
go get github.com/loov/goda
brew install dot # or equivalent
brew install redis # or equivalent
redis-server /usr/local/etc/redis.conf
```

### Boot server

```bash
go run main.go
```

Browse to http:/localhost:8888

## HTTP Endpoints

### Web

#### GET /
Content-Type: text/html; charset=utf-8
Example: http://localhost:8888
Example: https://gographs.io

#### GET /view?repo=[GO_REPO]
Content-Type: text/html; charset=utf-8
Example: http://localhost:8888/?repo=github.com/loov/goda?cluster=false
Example: https://gographs.io/?repo=github.com/siggy/gographs?cluster=false

#### GET /view?url=[SVG_URL]
Content-Type: text/html; charset=utf-8
Example: http://localhost:8888/?url=http://localhost:8888/?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg
Example: https://gographs.io/?url=http://localhost:8888/?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg

### API

#### GET /repo/[GITHUB_REPO].svg
Content-Type: image/svg+xml; charset=utf-8
Example: http://localhost:8888/repo/github.com/linkerd/linkerd2.svg
Example: https://gographs.io/repo/github.com/linkerd/linkerd2.svg

#### GET /repo/[GITHUB_REPO].dot
Content-Type: text/plain; charset=utf-8
Example:  http://localhost:8888/repo/github.com/linkerd/linkerd2.svg
Example:  https://gographs.io/repo/github.com/linkerd/linkerd2.svg


## TODO

- refresh button
- fix thumbnail resize/refresh jitter
- fix irc logo thumbnail border aspect ratio
- runtime flags
  - log-level
  - redis-server
- render fetch errors
- show progress during repo compile
- prevent normal dragging on mobile
- generic svg and dot support, paste and url
- permalink URLs
  - https://godoc.org/github.com/linkerd/linkerd2/pkg/k8s
- hide cluster checkbox when viewing raw svg
- consider capturing all mouse scrolling over every element
- firefox support
- remove or repurpose localhost URLs in readme
