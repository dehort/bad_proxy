package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type ReadWriteDiscarder struct {
	net.Conn
}

func (d *ReadWriteDiscarder) Write(p []byte) (n int, err error) {
	if d.Conn != nil {
        log.Println("Passing through a write")
		return d.Conn.Write(p)
	}

    log.Println("discarding a write")
	return len(p), nil
}

func (d *ReadWriteDiscarder) Read(p []byte) (n int, err error) {
	if d.Conn != nil {
        log.Println("Passing through a read")
		return d.Conn.Read(p)
	}

    log.Println("discarding a read")
	return 0, nil
}

func (d *ReadWriteDiscarder) Close() (err error) {
	if d.Conn != nil {
		retVal := d.Conn.Close()
		d.Conn = nil
		return retVal
	}

	return nil
}

func main() {
	l, err := net.Listen("tcp", "localhost:4242")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

        log.Println("Accepted a client connection")

		proxy(conn)
	}

}

func proxy(client net.Conn) {
	defer client.Close()

	server, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Opened a connection to the backend server")

	defer server.Close()

	server = &ReadWriteDiscarder{server}

	go io.Copy(server, client)
	go io.Copy(client, server)

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signalChan
	log.Println("Received signal to shutdown: ", sig)
	server.Close()

	signalChan = make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	sig = <-signalChan
}
