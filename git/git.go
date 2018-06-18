package git

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/faceair/betproxy"
)

var vscRegex = regexp.MustCompile(`([A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)+?)(/info/refs|/git-upload-pack|\?go-get=1)`)

func NewServer(gopath string) *Server {
	err := os.Setenv("GOPATH", gopath)
	if err != nil {
		panic(err)
	}

	g := &Server{
		gopath: gopath,
		queue:  make(chan string, 1024),
	}
	go g.updateLoop()
	return g
}

type Server struct {
	gopath string
	queue  chan string
	upTime sync.Map
}

func (g *Server) Do(req *http.Request) (*http.Response, error) {
	match := vscRegex.FindStringSubmatch(req.URL.String())
	if match == nil {
		return betproxy.HTTPError(http.StatusBadRequest, "url not match", req), nil
	}

	repoPath := match[1]
	urlPath := match[3]

	header := http.Header{
		"Expires":       []string{"Fri, 01 Jan 1980 00:00:00 GMT"},
		"Pragma":        []string{"no-cache"},
		"Cache-Control": []string{"no-cache, max-age=0, must-revalidate"},
	}
	switch urlPath {
	case "?go-get=1":
		html := fmt.Sprintf(`<meta name="go-import" content="%s git https://%s">`, repoPath, repoPath)
		return betproxy.HTTPText(http.StatusOK, nil, html, req), nil

	case "/info/refs":
		if !g.checkRepo(repoPath) {
			err := g.clone(repoPath)
			if err != nil {
				return nil, err
			}
		}

		select {
		case g.queue <- repoPath:
		default:
		}

		service := strings.Replace(req.FormValue("service"), "git-", "", 1)
		args := []string{service, "--stateless-rpc", "--advertise-refs", "."}
		refs, err := g.cmd(repoPath, args...).Output()
		if err != nil {
			return nil, err
		}
		serverAdvert := fmt.Sprintf("# service=git-%s\n", service)
		body := fmt.Sprintf("%04x%s0000%s", len(serverAdvert)+4, serverAdvert, refs)

		header.Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service))
		return betproxy.HTTPText(http.StatusOK, header, body, req), nil

	case "/git-upload-pack":
		args := []string{"upload-pack", "--stateless-rpc", "."}
		cmd := g.cmd(repoPath, args...)

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		err = cmd.Start()
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(stdin, req.Body)
		if err != nil {
			return nil, err
		}

		header.Set("Content-Type", "application/x-git-upload-pack-result")
		return betproxy.NewResponse(http.StatusOK, header, NewStdoutReader(stdout, cmd), req), stdin.Close()
	}

	return betproxy.HTTPError(http.StatusBadRequest, "url not match", req), nil
}

func (g *Server) clone(repoPath string) error {
	g.upTime.Store(repoPath, time.Now())

	logger := NewLogBuffer("Go Get")
	cmd := exec.Command("go", []string{"get", "-d", "-f", "-u", "-v", repoPath}...)
	cmd.Dir = g.gopath
	cmd.Stderr = logger
	err := cmd.Run()
	if err != nil {
		if strings.Contains(logger.String(), "no Go files") {
			err = nil
		} else {
			err = fmt.Errorf("%s %s", logger.String(), err.Error())
		}
	}
	return err
}

func (g *Server) updateLoop() {
	for {
		repoPath := <-g.queue
		if ut, ok := g.upTime.Load(repoPath); ok {
			if ut.(time.Time).Sub(time.Now()) < time.Minute {
				continue
			}
		}
		g.clone(repoPath)
	}
}

func (g *Server) checkRepo(dir string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/src/%s/.git", g.gopath, dir))
	return err == nil
}

func (g *Server) cmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = fmt.Sprintf("%s/src/%s", g.gopath, dir)
	return cmd
}

func NewStdoutReader(stdout io.Reader, cmd *exec.Cmd) io.Reader {
	return io.MultiReader(stdout, &StdoutCloser{cmd})
}

type StdoutCloser struct {
	cmd *exec.Cmd
}

func (c *StdoutCloser) Read(p []byte) (n int, err error) {
	if err = c.cmd.Wait(); err != nil {
		return 0, err
	}
	return 0, io.EOF
}

func NewLogBuffer(prefix string) *LogBuffer {
	return &LogBuffer{
		prefix: prefix,
		buffer: make([]byte, 0, 128),
	}
}

type LogBuffer struct {
	prefix string
	buffer []byte
}

func (l *LogBuffer) Write(p []byte) (n int, err error) {
	log.Printf("%s: %s", l.prefix, p)
	l.buffer = append(l.buffer, p...)
	return len(p), nil
}

func (l *LogBuffer) String() string {
	return string(l.buffer)
}
