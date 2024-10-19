# glover
Make a heatmap for my lovely Glove80.

## Run:
### Track
```shell
go build
./glover track -f /dev/tty.usbmodem12301 -f /dev/tty.usbmodem12401
```

### Show
TODO 

## Develop

### Live-reload
For continuous updates, it's easy to use [air](github.com/air-verse/air). I tested this repo with v1.61.1
```
air
```

### Tests
```
go test ./...
go test -bench -benchtime=10s ./...
```
