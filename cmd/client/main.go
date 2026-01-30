package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/armon/go-socks5"
)

type Config struct {
	ControlServer string `json:"server"`
	HTTPProxy     string `json:"http_proxy"`
	SOCKSProxy    string `json:"socks_proxy"`
	Autostart     bool   `json:"autostart"`
}

const BACKEND_ADDR = "mc.mrpickle.ca:8989"

type ProxyManager struct {
	httpListener  net.Listener
	socksListener *socks5.Server
	backendAddr   string

	running bool
}

func main() {
	fmt.Println("Loading from CLI, default configs used")
	fmt.Println("Address: " + BACKEND_ADDR)
	proxyManager := ProxyManager{}
	proxyManager.Start()
	for proxyManager.running {

	}
}

func (manager *ProxyManager) Start() error {
	if manager.running {
		return nil
	}
	manager.backendAddr = BACKEND_ADDR
	manager.running = true
	go func() {
		manager.httpListener = startHttps()
	}()
	go func() {
		manager.socksListener = startSocks5()

	}()

	return nil
}

func (manager *ProxyManager) Stop() error {
	if !manager.running {
		return nil
	}
	if manager.httpListener != nil {
		manager.httpListener.Close()
	}
	if manager.socksListener != nil {
		// figure out how to stop socks5 server
	}
	manager.running = false
	return nil
}

func startHttps() net.Listener {
	fmt.Println("client starting...")
	listener, err := net.Listen("tcp", ":8081")
	log.Println("HTTP Proxy listnening on :8081")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleHTTP(conn)
	}

	return listener

}

func startSocks5() *socks5.Server {
	conf := &socks5.Config{
		Dial: socksDial,
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SOCKS5 listening on 127.0.0.1:8082")
	err = server.ListenAndServe("tcp", "127.0.0.1:8082")
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func socksDial(ctx context.Context, network, addr string) (net.Conn, error) {
	// addr is "host:port"
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		InsecureSkipVerify: true, // only for testing/self-signed certs
	}

	serverConn, err := tls.Dial("tcp", BACKEND_ADDR, config)
	if err != nil {
		return nil, err
	}

	// Send your control protocol
	_, err = serverConn.Write([]byte(
		"SOCKS\n" + host + "\n" + port + "\n\n",
	))
	if err != nil {
		serverConn.Close()
		return nil, err
	}

	return serverConn, nil
}

func handleHTTP(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
		conn.Close()
	}()
	config := &tls.Config{
		InsecureSkipVerify: true, // only for testing/self-signed certs
	}

	reader := bufio.NewReader(conn)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	log.Println("First line recieved ", firstLine)
	if firstLine == "EOF" {
		return
	}

	list := strings.Split(firstLine, " ")
	method := list[0]
	hostPortCombo := strings.Split(list[1], ":")
	host := strings.TrimSpace(hostPortCombo[0])
	port := strings.TrimSpace(hostPortCombo[1])
	outbound, err := tls.Dial("tcp", BACKEND_ADDR, config)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		outbound.Write([]byte(fmt.Sprintf("HTTPS\n%s\n%s\n", host, port)))

	} else {
		outbound.Write([]byte(fmt.Sprintf("HTTP\n%s\n%s\n", host, port)))

	}
	go func() {
		_, _ = io.Copy(outbound, conn)
		conn.Close()
		outbound.Close()
	}()
	_, _ = io.Copy(conn, outbound)

}
