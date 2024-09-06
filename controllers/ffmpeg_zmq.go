package controllers

import (
	"fmt"
	"time"

	zmq "github.com/pebbe/zmq4"
)

const DO_ZMQ_CROSSFADES = true

var zmqAddresses = map[string]string{
	"fm":  "tcp://localhost:5555",
	"dab": "tcp://localhost:5556",
}

func zmq_crossfade(router, currentSource, newSource string) {
	if !DO_ZMQ_CROSSFADES {
		return
	}

	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()

	address := zmqAddresses[router]
	socket.Connect(address)

	for i := 1; i <= 5; i++ {
		fadeVolume(socket, currentSource, 5-i)
		fadeVolume(socket, newSource, i)

		time.Sleep(600 * time.Millisecond)
	}
}

func fadeVolume(socket *zmq.Socket, source string, level int) {
	msg := fmt.Sprintf("volume@s%s volume %.1f", source, float64(level)/5)
	socket.Send(msg, 0)
	socket.Recv(0)
}
