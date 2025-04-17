# glover

Make a heatmap for my lovely Glove80.

## Keyboard setup

In order for keyboard to report events, you need to put special value in config.
Refer to [ZMK docs](https://zmk.dev/docs/development/usb-logging#enabling-logging-on-older-boards)
for this. For example, in my `glove80.conf`, I have this line:

```conf
CONFIG_ZMK_USB_LOGGING=y
```

I am not aware of any way to enable similar feature of key logging without
a need for usb connection, so currently there's no other option other than
putting the line above to your keyboards' config and using it via usb.

## Run

### Track

```shell
make build
./tmp/glover track -f /dev/tty.usbmodem12301 -f /dev/tty.usbmodem12401 -o keypresses.sqlite -v
```

This will also open a web interface, default location is localhost:3000

Interface looks roughly like this:
![preview](img/preview.png)

### Show

In case if you don't need active key tracking, you can only run the web interface

```shell
./tmp/glover show -s keypresses.sqlite -p 8000
```

## Develop

### Live-reload

For development, it's easy to use [air](https://github.com/air-verse/air). Make sure to install npm to make changes
to the css/js related things.

```shell
make run-dev
```

### Tests

```shell
go test ./...
go test -bench -benchtime=10s ./...
```

### Lint

Make sure you have `prettier`, `tailwindcss` and `golangci-lint` installed.

```shell
brew install golangci-lint
npm install
```

```shell
make lint
```
