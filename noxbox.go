package main

import (
	"bytes"
	zmq "github.com/pebbe/zmq4"
	serial "github.com/tarm/goserial"
	"io"
	"log"
	"strings"
)

type NOXBOX struct {
	port    io.ReadWriteCloser
	site    string
  device  string
	address string
}

func (noxbox NOXBOX) Sample() string {
	c := serial.Config{Name: noxbox.device, Baud: 9600}
	port, err := serial.OpenPort(&c)
	noxbox.port = port
	if err != nil {
		log.Fatal(err)
	}
	return noxbox.read("\r")
}

func (noxbox NOXBOX) read(sep string) string {
	result := new(bytes.Buffer)

	query := noxbox.address + "no\r"
	_, err := noxbox.port.Write([]byte(query))
	if err != nil {
		log.Fatal(err)
	}

	for !strings.Contains(result.String(), sep) {
		buffer := make([]byte, 1024)
		n, err := noxbox.port.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}
		result.Write(buffer[:n])
	}
	return result.String()
}

func main() {
	noxbox := NOXBOX{}
	noxbox.site = "glbrc"
	noxbox.device = "/dev/ttyS5"
	noxbox.address = "\xAA"

	socket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	socket.Bind("tcp://*:5557")

	for {
		sample := noxbox.Sample()
		log.Print(sample)
		socket.Send(sample, 0)
	}
}
