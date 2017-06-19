package git

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_URL_UnsuccesfulURLParsing(t *testing.T) {
	var testCases = []struct {
		rawurl string
		err    error
	}{
		{rawurl: "   https://github.com  ", err: ErrURLIsWhitespacePadded},
		{rawurl: "git://github.com", err: ErrURLPathMissing},
		{rawurl: "git://git@github.com", err: ErrURLPathMissing},
		{rawurl: "git+ssh://github.com", err: ErrURLPathMissing},
	}
	for _, tc := range testCases {
		_, err := Parse(tc.rawurl)
		if err != tc.err {
			t.Fatalf("got error %s want %s", err, nil)
		}
	}
}

func Test_URL_SuccesfulURLParsing(t *testing.T) {
	var testCases = []struct {
		rawurl string
		gu     *url.URL
	}{

		// https:// URLs
		{rawurl: "https://github.com/path.git", gu: &url.URL{Host: "github.com", Path: "/path.git", Scheme: "https"}},
		{rawurl: "https://github.com:1234/path.git", gu: &url.URL{Host: "github.com:1234", Path: "/path.git", Scheme: "https"}},
		{rawurl: "https://user@github.com/path.git", gu: &url.URL{Host: "github.com", Path: "/path.git", Scheme: "https", User: url.User("user")}},
		{rawurl: "https://user:pw@github.com/path.git", gu: &url.URL{Host: "github.com", Path: "/path.git", Scheme: "https", User: url.UserPassword("user", "pw")}},

		// git+ssh:// URLs
		{rawurl: "git+ssh://git@192.168.50.4/path.git", gu: &url.URL{Host: "192.168.50.4", Path: "/path.git", Scheme: "git+ssh", User: url.User("git")}},
		{rawurl: "git+ssh://git@192.168.50.4:6789/path.git", gu: &url.URL{Host: "192.168.50.4:6789", Path: "/path.git", Scheme: "git+ssh", User: url.User("git")}},

		// ssh:// URLs (explicit)
		{rawurl: "ssh://git@69.73.176.35:7978/pedro/bridgeloannetwork.com.git", gu: &url.URL{Host: "69.73.176.35:7978", Path: "/pedro/bridgeloannetwork.com.git", Scheme: "ssh", User: url.User("git")}},

		// ssh:// URLs (implicit, aka scp-like (no port allowed))
		{rawurl: "192.168.50.4/path.git", gu: &url.URL{Host: "192.168.50.4", Path: "/path.git", Scheme: "ssh"}},
		{rawurl: "git@github.com/path.git", gu: &url.URL{Host: "github.com", Path: "/path.git", Scheme: "ssh", User: url.User("git")}},
		{rawurl: "git@192.168.50.4:7889", gu: &url.URL{Host: "192.168.50.4", Scheme: "ssh", User: url.User("git"), Path: "/7889"}},
		{rawurl: "git@192.168.50.4:user", gu: &url.URL{Host: "192.168.50.4", Scheme: "ssh", User: url.User("git"), Path: "/user"}},
		{rawurl: "git@69.73.176.35:7978/pedro/bridgeloannetwork.com.git", gu: &url.URL{Host: "69.73.176.35", Path: "/7978/pedro/bridgeloannetwork.com.git", Scheme: "ssh", User: url.User("git")}},
		// {rawurl: "git@192.168.50.4:/user", gu: &url.URL{Host: "192.168.50.4", Path: "//user", Scheme: "ssh", User: url.User("git")}},
	}
	for _, tc := range testCases {
		gu, err := Parse(tc.rawurl)
		if err != nil {
			t.Fatalf("got error:\n%s\nwant %v (rawurl: %v)", err, nil, tc)
		}
		if tc.gu == nil {
			continue // probably only checking err
		}
		if have, want := &gu.URL, tc.gu; !reflect.DeepEqual(want, have) {
			t.Fatalf("\n\texpected %#v\n\tgot      %#v", want, have)
		}
	}
}

func Test_URL_Copy(t *testing.T) {
	u1, _ := Parse("http://user:pass@example.com/path/to/repo.git")
	u2 := u1.Copy()
	u2.Host = "example.org"
	u2.User = nil
	if u1.Host == u2.Host {
		t.Fatalf("expected changes to u2 property not to affect u1")
	}
	if u1.User == u2.User {
		t.Fatalf("expected changes to u2 pointer member not to affect u1")
	}
}
