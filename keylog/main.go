package main

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

func main() {

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	for _, port := range ports {
		fmt.Printf("Port:%v\n", port)
	}

	port, err := serial.Open("/dev/tty.usbmodem12301", &serial.Mode{
		BaudRate: 9600,
	})
	defer func() {
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	port.SetReadTimeout(time.Second)

	if err != nil {
		log.Fatal(err)
	}

	buff := make([]byte, 1024)
	for {
		n, err := port.Read(buff)

		if err != nil {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}
		fmt.Printf("%v", string(buff[:n]))
	}

}
