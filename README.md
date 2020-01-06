## Local dev

### First-time setup

```bash
go get github.com/loov/goda
brew install dot # or equivalent
```

### Boot server

```bash
go run main.go
```

Browse to http:/localhost:8888

## API

GET /repo/[GITHUB_REPO].svg
```bash
curl /repo/github.com/linkerd/linkerd2.svg
```

PUT /repo/[GITHUB_REPO].svg
```bash
curl -X PUT http://localhost:8888/repo/github.com/linkerd/linkerd2.svg
```

## TODO

- fix cancelled svg request on repo change
  - don't request repo svg twice for main and thumb
- handle failed svg in svg-load
- cluster checkbox
- redis cache
- typeahead
- direct links for svg, png, dot
- generic svg and dot support, paste and url
