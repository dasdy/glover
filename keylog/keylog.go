package keylog

import (
	"errors"
	"log"

	"github.com/dasdy/glover/db"
	"github.com/dasdy/glover/keylog/parser"
)

func Loop(ch <-chan string, storage db.Storage, enableLogs bool) {
	for line := range ch {
		parsed, err := parser.ParseLine(line)
		if err != nil && !errors.Is(err, parser.ErrEmptyLine) {
			log.Printf("Got warning: %s\nline: '%s'", err.Error(), line)
		}

		if parsed != nil {
			if enableLogs {
				log.Printf("got keypress: %+v", *parsed)
			}

			if storage.Store(parsed) != nil {
				log.Printf("could not log item: %s", err.Error())
			}
		}
	}

	log.Println("Channel closed; bailing out")
}
