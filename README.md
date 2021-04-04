# gographs

[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/siggy/gographs?cluster=true)
[![Go Report Card](https://goreportcard.com/badge/github.com/siggy/gographs)](https://goreportcard.com/report/github.com/siggy/gographs)

[gographs](https://gographs.io) renders dependency graphs for Go packages.

[![gographs dependency example](https://gographs.io/graph/github.com/siggy/gographs.svg?cluster=true)](https://gographs.io)

## Badge Markdown

```md
[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/GO_REPO?[cluster=true|false])
```

Example
```md
[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/siggy/gographs?cluster=true)
```

## HTTP Endpoints

| Endpoint | Desc |
| --- | --- |
| [/](https://gographs.io) | Defaults to rendering this Go repo. |
| [/repo/GO_REPO?cluster=false\|true](https://gographs.io/repo/github.com/siggy/gographs?cluster=true) | Permalink to a repo. Use `POST` to refresh. |
| [/graph/GO_REPO.svg?cluster=false\|true](https://gographs.io/graph/github.com/siggy/gographs.svg?cluster=true) | SVG direct link. Use `POST` to refresh. |
| [/graph/GO_REPO.dot?cluster=false\|true](https://gographs.io/graph/github.com/siggy/gographs.dot?cluster=true) | GraphViz DOT direct link. Use `POST` to refresh. |
| [/svg?url=SVG_URL](https://gographs.io/svg?url=https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg) | Permalink to view an arbitrary SVG URL. |

## Local dev

### First-time setup

```bash
go install github.com/loov/goda@v0.4.1
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
deserve special mention, as this site is essentially a mashup of them:

- [goda](https://github.com/loov/goda)
- [SVGPan](https://github.com/bumbu/svg-pan-zoom)

[`pkg/repo`](./pkg/repo) is based on [Go Report Card](https://github.com/gojp/goreportcard)

[![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/siggy/gographs?cluster=true) courtesy of [shields.io](https://shields.io/)

## Support

It costs about one cup of coffee per day to run
[gographs.io](https://gographs.io). Buy me a cup of coffee and you've funded
[gographs.io](https://gographs.io) for a whole day!

<a href="https://www.buymeacoffee.com/siggy" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>

[![cash.app](public/img/cash.app.svg)](https://cash.app/$siggy)
[![paypal.me](public/img/paypal.me.svg)](https://paypal.me/andrewseigner)
