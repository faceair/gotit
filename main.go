package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/faceair/betproxy"
	"github.com/faceair/betproxy/mitm"
	"github.com/faceair/gotit/git"
)

func main() {
	var host, capath, gopath string
	var port int
	flag.StringVar(&capath, "capath", os.TempDir(), "path to save gotit certificate")
	flag.StringVar(&gopath, "gopath", os.Getenv("GOPATH"), "path to save git repository")
	flag.StringVar(&host, "host", "0.0.0.0", "proxy server host")
	flag.IntVar(&port, "port", 0, "proxy server port")
	flag.Parse()

	if port == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	cacert, cakey, err := loadCA(capath)
	if err != nil {
		panic(err)
	}
	tlsCfg, err := mitm.NewConfig(cacert, cakey)
	if err != nil {
		panic(err)
	}

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

func loadCA(capath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certpath := path.Join(capath, "gotit.cert.pem")
	keypath := path.Join(capath, "gotit.key.pem")

	if _, err := os.Stat(certpath); os.IsNotExist(err) {
		err := generateCA(capath, certpath, keypath)
		if err != nil {
			return nil, nil, err
		}
	}

	cert, err := ioutil.ReadFile(certpath)
	if err != nil {
		return nil, nil, err
	}
	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, nil, err
	}
	certBlock, _ := pem.Decode(cert)
	if certBlock == nil {
		return nil, nil, errors.New("Failed to decode ca certificate")
	}
	keyBlock, _ := pem.Decode(key)
	if keyBlock == nil {
		return nil, nil, errors.New("Failed to decode ca private key")
	}
	rawKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	rawCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return rawCert, rawKey, nil
}

func generateCA(capath, certpath, keypath string) error {
	err := os.MkdirAll(capath, os.ModePerm)
	if err != nil {
		return err
	}

	cacert, cakey, err := mitm.NewAuthority("gotit", "faceair", 10*365*24*time.Hour)
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, cacert, cacert, &cakey.PublicKey, cakey)
	if err != nil {
		return err
	}
	certOut, err := os.Create(certpath)
	if err != nil {
		return err
	}
	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err = certOut.Close(); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keypath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(cakey)}); err != nil {
		return err
	}

	return keyOut.Close()
}
