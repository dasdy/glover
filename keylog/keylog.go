package keylog

import (
	"glover/db"
	"glover/keylog/parser"
	"log"
)

func KeyLogLoop(ch <-chan string, done <-chan bool, storage db.Storage, enableLogs bool) {
out:
	for {
		select {
		case line := <-ch:
			parsed, err := parser.ParseLine(line)
			if err != nil {
				log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
			}

			if parsed != nil {
				if enableLogs {
					log.Printf("Event! %v", *parsed)
				}
				storage.Store(parsed)
			}
		case <-done:
			log.Println("Received done from port readers, bailing out")
			break out
		}
	}
}
