package main

import (
	"bytes"
	/* "encoding/json" */
	zmq "github.com/pebbe/zmq4"
	serial "github.com/tarm/goserial"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NOXBOX struct {
	site    string
	device  string
	address string
}

type Message struct {
	nox  float64
	site string
	at time.Time
}

func (noxbox NOXBOX) Sample() string {

	data := noxbox.read()
	for data == "" {
		data = noxbox.read()
	}
	nox := noxbox.parse(data)

	/* message := Message{nox:nox, site: noxbox.site, at: time.Now()} */
	/* json_message, err := json.Marshal(message) */
	/* if err != nil { */
	/* 	log.Fatal(err) */
	/* } */
	return string(strconv.FormatFloat(nox,'f',-1, 64))
}

func (noxbox NOXBOX) parse(data string) float64 {
	r, _ := regexp.Compile(`\d+E-?\d+`)
	no := r.FindString(data)
	result , err := strconv.ParseFloat(no, 64)
  if err != nil { log.Fatal(err) }
  return result
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
