package main

import (
	"fmt"
	"log"
	"os"
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

	args := os.Args

	var fname string
	if len(args) < 2 {
		fname = "/dev/tty.usbmodem12301"
	} else {
		fname = args[1]
	}

	port, err := serial.Open(fname, &serial.Mode{
		BaudRate: 9600,
	})
	defer func() {
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	port.SetReadTimeout(10 * time.Second)

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

	fmt.Println("Read ended")
}
