package git

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func NewTestServer() *Server {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	return NewServer(dir)
}

func TestURLNotMatch(t *testing.T) {
	server := NewTestServer()
	defer os.RemoveAll(server.gopath)

	u, err := url.Parse("https://faceair.hi/gotit")
	if err != nil {
		t.Error("parse url failed")
	}
	res, err := server.Do(&http.Request{
		URL: u,
	})
	if err != nil {
		t.Error("request failed")
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Error("url should not matched!")
	}
}

func TestMetadata(t *testing.T) {
	server := NewTestServer()
	defer os.RemoveAll(server.gopath)

	u, err := url.Parse("https://faceair.hi/gotit?go-get=1")
	if err != nil {
		t.Error("parse url failed")
	}
	res, err := server.Do(&http.Request{
		URL: u,
	})
	if err != nil {
		t.Error("request failed")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("read body failed")
	}
	if !strings.Contains(string(body), `<meta name="go-import" content="faceair.hi/gotit git https://faceair.hi/gotit">`) {
		t.Error("read metadata failed")
	}
}

func TestGitClone(t *testing.T) {
	server := NewTestServer()
	defer os.RemoveAll(server.gopath)

	u, err := url.Parse("https://github.com/faceair/betproxy/info/refs?service=git-upload-pack")
	if err != nil {
		t.Error("parse url failed")
	}

	t1 := time.Now()
	res, err := server.Do(&http.Request{
		URL: u,
	})
	t2 := time.Now().Sub(t1)
	if err != nil {
		t.Error("request failed")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("read body failed")
	}
	if !strings.Contains(string(body), `service=git-upload-pack`) {
		t.Error("read info/refs failed")
	}

	t3 := time.Now()
	res, err = server.Do(&http.Request{
		URL: u,
	})
	t4 := time.Now().Sub(t3)

	if err != nil {
		t.Error("request failed")
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("read body failed")
	}
	if !strings.Contains(string(body), `service=git-upload-pack`) {
		t.Error("read info/refs failed")
	}

	if t4 > t2 {
		t.Error("cache miss")
	}

	u, err = url.Parse("https://github.com/faceair/betproxy/git-upload-pack")
	if err != nil {
		t.Error("parse url failed")
	}
	res, err = server.Do(&http.Request{
		URL:  u,
		Body: ioutil.NopCloser(strings.NewReader("00aawant 0a6c1534ac051e00264e718d342aed48b1af197d multi_ack_detailed no-done side-band-64k thin-pack ofs-delta deepen-since deepen-not agent=git/2.15.2.(Apple.Git-101.1)\n00000009done\n")),
	})
	if err != nil {
		t.Error("request failed")
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("read body failed")
	}
	if !strings.Contains(string(body), `Counting objects`) {
		t.Error("clone repository failed")
	}
}

func TestLogBuffer(t *testing.T) {
	log := NewLogBuffer("git")
	log.Write([]byte("hi"))
	log.Write([]byte(","))
	log.Write([]byte("faceair"))
	if log.String() != "hi,faceair" {
		t.Error("log have no buffer")
	}
}
