package main

import (
	"bufio"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	serial "github.com/tarm/goserial"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"
)

type NOXBOX struct {
	site    string
	device  string
	address string
}

type Message struct {
	NOX  float64   `json:"nox"`
	Site string    `json:"site"`
	At   time.Time `json:"at"`
}

func (noxbox NOXBOX) Sample() (string, error) {
	c := serial.Config{Name: noxbox.device, Baud: 9600}
	port, err := serial.OpenPort(&c)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()
	data := noxbox.read(port)
	for data == "" {
		data = noxbox.read(port)
	}
	nox, err := noxbox.parse(data)
	if err != nil {
		return "", err
	}

	message := Message{nox, noxbox.site, time.Now()}
	json_message, err := json.Marshal(message)
	return string(json_message), err
}
func (noxbox NOXBOX) parse(data string) (float64, error) {
	/* log.Print(data) */
	r, _ := regexp.Compile(`\d+E-?\d+`)
	no := r.FindString(data)
	result, err := strconv.ParseFloat(no, 64)
	if err != nil {
		log.Print(err, data, no)
		return 0, err
	}
	return result, nil
}

func (noxbox NOXBOX) read(port io.ReadWriteCloser) string {

	query := noxbox.address + "no\r"
	_, err := port.Write([]byte(query))
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(port)
	reply, err := reader.ReadBytes('\x0A')
	if err != nil {
		if err == io.EOF {
			/* log.Print(err) */
		} else {
			log.Fatal(err)
		}
	}
	return string(reply)
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
		sample, _ := noxbox.Sample()
		if err == nil {
			log.Print(sample)
			socket.Send(sample, 0)
		}
		time.Sleep(10 * time.Second)
	}
}
