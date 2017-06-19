package git

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrURLIsWhitespacePadded = fmt.Errorf("git: url contains trailing or leading whitespace")
	ErrURLHostMissing        = fmt.Errorf("git: url host missing")
	ErrURLPathMissing        = fmt.Errorf("git: url path missing")
	ErrURLSchemeUnsupported  = fmt.Errorf("git: url scheme unsupported")
)

type URL struct {
	url.URL
}

func (self *URL) Copy() *URL {
	// Don't check error as we guard getting a good URL
	// via `NewRepository`, shouldn't be possible to make
	// it unparsable.
	cpy, _ := Parse(self.URL.String())
	return cpy
}

func (self *URL) String() string {
	return self.URL.String()
}

func Parse(rawurl string) (*URL, error) {

	if strings.TrimSpace(rawurl) != rawurl {
		return nil, ErrURLIsWhitespacePadded
	}

	var schemeInferred bool

	var firstColonPos int = strings.IndexRune(rawurl, ':')
	var firstSlashPos int = strings.IndexRune(rawurl, '/')

	var hasColon bool = firstColonPos > -1
	var hasSlash bool = firstSlashPos > -1

	u, err := url.Parse(rawurl)
	if err != nil {
		// url.Error sets the Op to parse here
		// https://github.com/golang/go/blob/476f55fd8aec65cb1bd3417dd1c8c583c9e385d8/src/net/url/url.go#L436
		urlErr, _ := err.(*url.Error)
		if strings.Compare(urlErr.Op, "parse") == 0 {
			// If scheme is unknown, assume ssh:// and remember it, it becomes
			// important in some cases below.
			schemeInferred = true
			// Re-parsing the URL here guards against the case that `git@github.com`
			// is assumed to be a URL{Host:"", Path:"git@github.com"}
			u, _ = url.Parse(fmt.Sprintf("ssh://%s", rawurl))
		} else {
			return nil, err
		}
	}

	// Check for scp-like URLs, we might have u.Scheme="ssh", but we know if we
	// inferred it, or were given it, so we ignore u.Scheme here.
	//
	//  * Scheme must have been *inferred* as SSH
	//  * A colon appears before the first slash in the hostname + path part of the URL
	//  * If a colon is found, we must also find a slash, else it's undefined (citation needed)
	//
	// From the git Docs [1]
	//	>	An alternative scp-like syntax may also be used with the ssh protocol:
	// 	>	[user@]host.xz:path/to/repo.git/
	// 	>	This syntax is only recognized if there are no slashes before the first colon.
	//
	// [1]: https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a
	if schemeInferred {
		// Is this really fatal, or can we let '$ git' handle it?
		if (hasColon && !hasSlash) || (firstColonPos < firstSlashPos) {
			rawurl = strings.Replace(rawurl, ":", "/", 1)
			return Parse(fmt.Sprintf("ssh://%s", rawurl))
		}
	}

	// See https://github.com/golang/go/issues/18824
	if len(u.Host) == 0 {
		if strings.HasPrefix(rawurl, "ssh://") {
			return nil, ErrURLHostMissing
		} else {
			return Parse(fmt.Sprintf("ssh://%s", rawurl))
		}
	}

	if len(u.Path) == 0 {
		return nil, ErrURLPathMissing
	}

	switch u.Scheme {
	case "ssh", "http", "https", "git+ssh", "":
		break
	default:
		return nil, ErrURLSchemeUnsupported
	}

	return &URL{URL: *u}, nil
}
