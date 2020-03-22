package main

import (
	"flag"

	ssh "github.com/sarovkalach/go_ssh"
	log "github.com/sirupsen/logrus"
)

func parseFlags() map[string]string {
	file := flag.String("file", "proxy.txt", "file to process")
	username := flag.String("u", "root", "file to process")
	nThreads := flag.String("n", "128", "N threads")
	timeout := flag.String("t", "5", "request timeout")
	host := flag.String("h", "192.168.0.1", "host")
	flag.Parse()

	return map[string]string{
		"filename": *file,
		"nThreads": *nThreads,
		"timeout":  *timeout,
		"host":     *host,
		"username": *username,
	}
}

func main() {
	cfg := parseFlags()
	log.SetLevel(log.DebugLevel)
	c := ssh.NewConnector(cfg)

	go c.Start()
	log.Info(<-c.StopCh)
}
