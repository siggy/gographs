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

- ensure multiple calls to addEventListener after each svg load are ok
- redis cache
  - cache intermediate steps
- typeahead
- direct links for svg, png, dot
- generic svg and dot support, paste and url
- cluster checkbox should auto-reload already-displaying repos
- hideable control panel and thumbnail browser
- list most popular repos / queries
