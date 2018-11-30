package git

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

var urlRegex = regexp.MustCompile(`([A-Za-z0-9_.-]+((/[A-Za-z0-9_.-]+)+?)?)/?(/info/refs|/git-upload-pack|\?go-get=1)`)
var repoRegex = regexp.MustCompile(`content="(.+?)\s+git\s+(.+)?"`)

// NewServer create a Server instance
// The gopath should be a valid folder and will store git repositories later
func NewServer(gopath string) *Server {
	err := os.Setenv("GOPATH", gopath)
	if err != nil {
		panic(err)
	}

	g := &Server{
		gopath: gopath,
		queue:  make(chan *cloneTask, 1024),
	}
	go g.cloneLoop()
	return g
}

// Server implement interface of betproxy.Client
type Server struct {
	gopath string
	queue  chan *cloneTask
	upTime sync.Map
}

// Do receive client requests and return git repository information
func (g *Server) Do(req *http.Request) (*http.Response, error) {
	match := urlRegex.FindStringSubmatch(req.URL.String())
	if match == nil {
		return HTTPRedirect("https://github.com/faceair/gotit", req), nil
	}

	repoPath := match[1]
	urlPath := match[4]

	header := http.Header{
		"Expires":       []string{"Fri, 01 Jan 1980 00:00:00 GMT"},
		"Pragma":        []string{"no-cache"},
		"Cache-Control": []string{"no-cache, max-age=0, must-revalidate"},
	}
	switch urlPath {
	case "?go-get=1":
		repoPath, err := g.getVSCRoot(repoPath)
		if err != nil {
			return nil, err
		}

		html := fmt.Sprintf(`<meta name="go-import" content="%s git https://%s">`, repoPath, repoPath)
		return betproxy.HTTPText(http.StatusOK, nil, html, req), nil

	case "/info/refs":
		if !g.checkRepo(repoPath) {
			task := newCloneTask(repoPath)
			g.queue <- task
			<-task.Done()
		} else {
			select {
			case g.queue <- newCloneTask(repoPath):
			default:
			}
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

	return nil, errors.New("url not match")
}

func (g *Server) cloneLoop() {
	for {
		task := <-g.queue
		if g.shouldUpdate(task.repoPath) {
			if err := g.clone(task.repoPath); err != nil {
				log.Printf("Clone Failed: %s", err.Error())
			}
		}
		close(task.Done())
	}
}

func (g *Server) shouldUpdate(repoPath string) bool {
	now := time.Now()
	if ut, ok := g.upTime.Load(repoPath); ok {
		if now.Sub(ut.(time.Time)) < time.Hour {
			return false
		}
	}
	g.upTime.Store(repoPath, now)
	return true
}

func (g *Server) clone(repoPath string) error {
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

func (g *Server) getVSCRoot(repoPath string) (string, error) {
	dirs := strings.Split(repoPath, "/")
	for len(dirs) > 0 {
		guessPath := strings.Join(dirs, "/")
		if g.checkRepo(guessPath) {
			return guessPath, nil
		}
		dirs = dirs[:len(dirs)-1]
	}
	res, err := http.Get(fmt.Sprintf("https://%s?go-get=1", repoPath))
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	matches := repoRegex.FindStringSubmatch(string(body))
	if len(matches) == 3 {
		return matches[1], nil
	}
	return "", errors.New("parse meta tags failed")
}

func (g *Server) checkRepo(repoPath string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/src/%s/.git", g.gopath, repoPath))
	return err == nil
}

func (g *Server) cmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = fmt.Sprintf("%s/src/%s", g.gopath, dir)
	return cmd
}

// NewStdoutReader return an io.reader that closes the command when it finishes reading data
func NewStdoutReader(stdout io.Reader, cmd *exec.Cmd) io.Reader {
	return io.MultiReader(stdout, &StdoutCloser{cmd})
}

// StdoutCloser implement interface of io.Reader
type StdoutCloser struct {
	cmd *exec.Cmd
}

// Read wait and close the child process to avoid zombie process
func (c *StdoutCloser) Read(p []byte) (n int, err error) {
	if err = c.cmd.Wait(); err != nil {
		return 0, err
	}
	return 0, io.EOF
}

// NewLogBuffer create a LogBuffer instance
func NewLogBuffer(prefix string) *LogBuffer {
	return &LogBuffer{
		prefix: prefix,
		buffer: make([]byte, 0, 128),
	}
}

// LogBuffer implement interface of io.Writer
type LogBuffer struct {
	prefix string
	buffer []byte
}

// Write write log to stdout and collect logs to buffer
func (l *LogBuffer) Write(p []byte) (n int, err error) {
	log.Printf("%s: %s", l.prefix, p)
	l.buffer = append(l.buffer, p...)
	return len(p), nil
}

// String return cached logs
func (l *LogBuffer) String() string {
	return string(l.buffer)
}

func newCloneTask(repoPath string) *cloneTask {
	return &cloneTask{
		repoPath: repoPath,
		done:     make(chan struct{}),
	}
}

type cloneTask struct {
	repoPath string
	done     chan struct{}
}

func (t *cloneTask) Done() chan struct{} {
	return t.done
}

// HTTPRedirect create a temporary redirect http.Response with giving url
func HTTPRedirect(url string, req *http.Request) *http.Response {
	res := betproxy.NewResponse(http.StatusTemporaryRedirect, http.Header{
		"Location": []string{url},
	}, nil, req)
	res.ContentLength = 0
	return res
}
