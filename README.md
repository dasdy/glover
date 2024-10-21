# glover

Make a heatmap for my lovely Glove80.

## Run:

### Track

```shell
make build
./bin/glover track -f /dev/tty.usbmodem12301 -f /dev/tty.usbmodem12401 -o keypresses.sqlite -v
```

This will also open a web interface, default location is localhost:3000

Interface looks roughly like this:
![preview](img/preview.png)

### Show

In case if you don't need active key tracking, you can only run the web interface

```shell
./bin/glover show -s keypresses.sqlite -p 8000
```

## Develop

### Live-reload

For continuous updates, it's easy to use [air](github.com/air-verse/air). I tested this repo with v1.61.1

```
make run-dev
```

### Tailwind

For active development you'll need a [tailwind-cli](https://tailwindcss.com/blog/standalone-cli) somewhere in your PATH.
To re-generate css if you are actively working on html templates, run

```shell
make tailwind-watch
```

### Tests

```
go test ./...
go test -bench -benchtime=10s ./...
```
