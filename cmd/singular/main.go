package main

import (
	"crypto/tls"
	"flag"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Dreamacro/singular"
	log "github.com/Sirupsen/logrus"
)

var (
	configPath = flag.String("config", "config.yml", "config file")
	cert       = flag.String("cert", "cert.pem", "cert file")
	key        = flag.String("key", "cert.key", "cert key file")
	useTLS     = flag.Bool("tls", false, "use TLS")
	logPath    = flag.String("log", "", "log path")

	retryTimes = 1000

	// DEBUG open debug mode
	DEBUG = flag.Bool("debug", false, "debug mode")
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
			log.Fatal("Open Log File Fail")
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stderr)
	}
	defer f.Close()

	config, err := singular.ParseConfig(*configPath)
	singular.PassOrFatal("Read Config Error", err)

	if *DEBUG {
		go func() {
			c := time.NewTicker(5 * time.Second).C
			for {
				<-c
				log.Infof("Goroutine Num: %d", runtime.NumGoroutine())
			}
		}()
	}

	var tlsConfig *tls.Config
	if *useTLS {
		var err error
		tlsConfig, err = singular.NewTLSConfig(*cert, *key)
		tlsConfig.InsecureSkipVerify = true
		if err != nil {
			log.Fatal("cert file or cert key file error")
		}
	}

	var wg sync.WaitGroup
	daemon := func(name, localAddr string) {
		rt := retryTimes
		client := singular.NewClient(name, localAddr, config.ServerAddr, *useTLS, tlsConfig)
		for rt != 0 {
			if err := client.Connect(); err != nil {
				rt--
				log.Printf("%s Retry after 5 second...", name)
				time.Sleep(5 * time.Second)
				continue
			}
		}
		wg.Done()
	}

	for name, addr := range config.Proxy {
		wg.Add(1)
		go daemon(name, addr)
	}
	wg.Wait()
}
