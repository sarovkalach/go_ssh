package connector

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	ssh "golang.org/x/crypto/ssh"
)

const buffSize = 15000000

type Connector struct {
	passCh      chan string
	username    string
	host        string
	threadCount int
	StopCh      chan string
	timeout     time.Duration
	wg          *sync.WaitGroup
}

func NewConnector(cfg map[string]string) *Connector {
	nThreads, _ := strconv.Atoi(cfg["nThreads"])
	c := &Connector{
		username:    cfg["username"],
		host:        cfg["host"] + ":22",
		threadCount: nThreads,
		timeout:     time.Duration(1 * time.Second),
		passCh:      make(chan string, buffSize),
		StopCh:      make(chan string),
		wg:          &sync.WaitGroup{},
	}
	c.readFile(cfg["filename"])

	return c
}

func (c *Connector) readFile(filename string) {
	readFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	for fileScanner.Scan() {
		text := fileScanner.Text()
		c.passCh <- text
	}

	log.Debug(fmt.Sprintf("Len of proxylist: %d", len(c.passCh)))
}

func (c *Connector) Start() {
	defer close(c.passCh)

	for {
		select {
		case <-c.StopCh:
			return
		default:
			for i := 0; i < c.threadCount; i++ {
				c.wg.Add(1)
				go c.connect(<-c.passCh)
				time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
			}
			c.wg.Wait()
		}
	}
}

func (c *Connector) connect(password string) {
	defer c.wg.Done()

	sshConfig := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		Timeout:         c.timeout,
	}

	conn, err := ssh.Dial("tcp", c.host, sshConfig)
	if err != nil {
		if !strings.Contains(err.Error(), "no supported methods remain") {
			c.passCh <- password
			return
		}
		log.Debug(err.Error(), "|", c.username+":"+password)
		return
	}
	defer conn.Close()

	c.StopCh <- c.username + ":" + password
}
