package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"
)

var PORT = ":8080"

func main() {

	log.Println("Server starting")
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	listner, error := tls.Listen("tcp4", PORT, config)

	if error != nil {
		log.Fatal(error)
	}
	defer listner.Close()

	for {
		connection, error := listner.Accept()
		if error != nil {
			log.Fatal(error)
			return
		}
		go handleConnection(connection)

	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	log.Println("Connection recieved from ", connection.RemoteAddr().String(), "\n")

	reader := bufio.NewReader(connection)

	//parse method
	method, _ := GrabLine(reader)

	host, _ := GrabLine(reader)
	port, _ := GrabLine(reader)

	if strings.Contains(host, "CONNECT") {
		//handle HTTPS
		list := strings.Split(host, " ")
		host = strings.Split(list[1], ":")[0]
		port = strings.Split(list[1], ":")[1]
		//protocol := list[2]

	}
	log.Println("Host: ", host, " Port: ", port)

	addr := host + ":" + port
	destinationConnection, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("Error ", err)
		connection.Write([]byte("Error\n"))
		connection.Close()
		return
	}
	defer destinationConnection.Close()
	if method == "HTTPS" {
		//handle HTTPS routing
		log.Println("HTTPS Routing being used")
		connection.Write([]byte("HTTPS/1.1 200 Connection Established\n"))

	} else if method == "SOCKS5" {
		connection.Write([]byte("OK\n"))
	} else {
		log.Println("HTTP Routing being used")

		connection.Write([]byte("HTTP/1.1 200 Connection Established\n"))
	}

	go func() {
		_, _ = io.Copy(destinationConnection, connection)
	}()
	_, _ = io.Copy(connection, destinationConnection)
}

func GrabLine(reader *bufio.Reader) (string, error) {
	line, error := reader.ReadString('\n')
	if error != nil {
		log.Fatal("Error reading host: ", error)
		return "Err", error
	}
	return strings.TrimSpace(line), nil
}
