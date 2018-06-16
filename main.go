package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faceair/betproxy"
	"github.com/faceair/betproxy/mitm"
	"github.com/faceair/gotit/git"
)

func main() {
	var host, gopath string
	var port int
	flag.StringVar(&gopath, "gopath", os.Getenv("GOPATH"), "path to the git repository")
	flag.StringVar(&host, "host", "0.0.0.0", "proxy server host")
	flag.IntVar(&port, "port", 0, "proxy server port")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	tlsCfg, err := generateTLSCfg()
	if err != nil {
		panic(err)
	}
	service, err := betproxy.NewService(fmt.Sprintf("%s:%d", host, port), tlsCfg)
	if err != nil {
		panic(err)
	}
	service.SetClient(git.NewServer(gopath))

	log.Fatal(service.Listen())
}

func generateTLSCfg() (*mitm.Config, error) {
	cacert, cakey, err := mitm.NewAuthority("betproxy", "betproxy", 10*365*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return mitm.NewConfig(cacert, cakey)
}
