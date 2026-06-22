package graph

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gogitclient "github.com/go-git/go-git/v5/plumbing/transport/client"
	gogithttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/net/html"
)

// toDir resolves a Go import path to a git clone URL and shallow-clones it into
// a fresh temp directory, returning the directory path.
func toDir(repo string) (string, error) {
	cloneURL, err := resolveGitURL(trimScheme(repo))
	if err != nil {
		return "", err
	}

	codeDir, err := os.MkdirTemp("", "") // already 0700 and empty
	if err != nil {
		return "", err
	}

	if err := gitClone(cloneURL, codeDir); err != nil {
		os.RemoveAll(codeDir)
		return "", err
	}
	return codeDir, nil
}

// resolveGitURL maps a Go import path to an https git clone URL.
func resolveGitURL(importPath string) (string, error) {
	// Fast path: well-known git hosts map directly to host/owner/repo.
	for _, host := range []string{"github.com/", "gitlab.com/", "bitbucket.org/"} {
		if strings.HasPrefix(importPath, host) {
			p := strings.SplitN(importPath, "/", 4)
			if len(p) < 3 || p[1] == "" || p[2] == "" {
				return "", fmt.Errorf("invalid repo path: %q", importPath)
			}
			return "https://" + strings.Join(p[:3], "/"), nil
		}
	}
	// Vanity path: resolve via ?go-get=1 meta tag.
	return resolveVanity(importPath)
}

// resolveVanity fetches https://<path>?go-get=1 and returns the git repo root
// advertised by the go-import meta tag. https + git only, SSRF-guarded.
func resolveVanity(importPath string) (string, error) {
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{DialContext: (&net.Dialer{Control: blockPrivate}).DialContext},
	}
	resp, err := client.Get("https://" + importPath + "?go-get=1")
	if err != nil {
		return "", fmt.Errorf("go-get lookup failed: %w", err)
	}
	defer resp.Body.Close()

	repo, err := parseGoImport(resp.Body, importPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(repo, "https://") {
		return "", fmt.Errorf("refusing non-https repo root: %q", repo)
	}
	return repo, nil
}

// parseGoImport returns the git repo root from the first matching
// <meta name="go-import" content="prefix git repo"> tag.
func parseGoImport(r io.Reader, importPath string) (string, error) {
	z := html.NewTokenizer(r)
	for {
		switch z.Next() {
		case html.ErrorToken:
			return "", fmt.Errorf("no git go-import meta tag for %q", importPath)
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data != "meta" {
				continue
			}
			var name, content string
			for _, a := range t.Attr {
				switch a.Key {
				case "name":
					name = a.Val
				case "content":
					content = a.Val
				}
			}
			if name != "go-import" {
				continue
			}
			// content == "import-prefix vcs repo-root"
			if f := strings.Fields(content); len(f) == 3 && f[1] == "git" &&
				strings.HasPrefix(importPath+"/", f[0]+"/") {
				return f[2], nil
			}
		}
	}
}

// blockPrivate rejects dialing non-public IPs (called per resolved address, so
// it also covers redirects).
func blockPrivate(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
		return fmt.Errorf("blocked non-public address: %s", address)
	}
	return nil
}

// installHTTPOnce registers the SSRF-protected HTTP client with go-git once.
var installHTTPOnce sync.Once

// gitClone shallow-clones cloneURL into dir using pure-Go git (no git binary required).
func gitClone(cloneURL, dir string) error {
	installHTTPOnce.Do(func() {
		gogitclient.InstallProtocol("https", gogithttp.NewClient(&http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{Control: blockPrivate}).DialContext,
			},
		}))
	})

	_, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
		URL:          cloneURL,
		Depth:        1,
		SingleBranch: true,
		Tags:         gogit.NoTags,
	})
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

// trimScheme strips a leading scheme (e.g. "https://") from a repo path.
func trimScheme(repo string) string {
	if _, after, ok := strings.Cut(repo, "://"); ok {
		return after
	}
	return repo
}
