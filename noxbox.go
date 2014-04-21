package main

import (
	"bytes"
	zmq "github.com/pebbe/zmq4"
	serial "github.com/tarm/goserial"
	/* "encoding/json" */
	"io"
	"log"
	"regexp"
	"strings"
	"time"
)

type NOXBOX struct {
	site    string
	device  string
	address string
}

type Message struct {
	no   string
	site string
}

func (noxbox NOXBOX) Sample() string {

	data := noxbox.read()
	nox := noxbox.parse(data)

	return nox
	/* m := Message(nox, noxbox.site) */
	/* return json.Marshal(m) */
}

func (noxbox NOXBOX) parse(result string) string {
	r, _ := regexp.Compile(`no (\d+E-?\d+)`)
	no := r.FindString(result)
	return no
}

func (noxbox NOXBOX) read() string {
	c := serial.Config{Name: noxbox.device, Baud: 9600}
	port, err := serial.OpenPort(&c)
	if err != nil {
		log.Fatal(err)
	}

	defer port.Close()

	result := new(bytes.Buffer)

	query := noxbox.address + "no\r"
	_, err = port.Write([]byte(query))
	if err != nil {
		log.Fatal(err)
	}

	for !strings.Contains(result.String(), "\r") {
		buffer := make([]byte, 1024)
		n, err := port.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
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
		time.Sleep(10)
	}
}
