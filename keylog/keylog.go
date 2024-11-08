package keylog

import (
	"log"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/keylog/parser"
)

func KeyLogLoop(ch <-chan string, storage db.Storage, enableLogs bool) {
	for line := range ch {
		parsed, err := parser.ParseLine(line)
		if err != nil {
			log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
		}

		if parsed != nil {
			if enableLogs {
				log.Printf("Event! %v", *parsed)
			}
			err := storage.Store(parsed)
			if err != nil {
				log.Printf("Could not log item: %s", err.Error())
			}
		}
	}

	log.Println("Channel closed; bailing out")
}
