package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"os"

	"github.com/Dreamacro/singular"
	log "github.com/Sirupsen/logrus"
)

var (
	port    = flag.Int("port", 8080, "listen port")
	cert    = flag.String("cert", "cert.pem", "cert file")
	key     = flag.String("key", "cert.key", "cert key file")
	useTLS  = flag.Bool("tls", false, "use TLS")
	logPath = flag.String("log", "", "log path")
)

func init() {
	// log config
	log.SetFormatter(&log.TextFormatter{})

	// flag
	flag.Parse()
}

func main() {
	var f *os.File
	var err error
	if *logPath != "" {
		f, err = os.OpenFile(*logPath, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal("Open Log Fail")
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stderr)
	}
	defer f.Close()

	var tlsConfig *tls.Config
	if *useTLS {
		var err error
		tlsConfig, err = singular.NewTLSConfig(*cert, *key)
		tlsConfig.Rand = rand.Reader
		if err != nil {
			log.Fatal("cert file or cert key file error")
		}
	}

	server := singular.NewServer(*useTLS, tlsConfig)
	server.Serve(*port)
}
