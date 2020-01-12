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

- list most popular repos / queries
- refresh button
- fix irc logo thumbnail border aspect ratio
- runtime flags
  - log-level
  - redis-server
- render fetch errors
- show progress during repo compile
- prevent normal dragging on mobile
- ensure multiple calls to addEventListener after each svg load are ok
- typeahead
- direct links for png, dot
- generic svg and dot support, paste and url
- permalink URLs
  - https://godoc.org/github.com/linkerd/linkerd2/pkg/k8s
- hide cluster checkbox when viewing raw svg
- consider capturing all mouse scrolling over every element
- firefox support
