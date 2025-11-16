# glover

- Record keypresses from Glove80 (or potentially other ZMK-based keyboards)
- Multiple ways to collect the data:
  - Through explicit pointing to usb devices
  - Parse logs from standard input
  - Search and auto-connect to devices - either as a one-time operation, or as a
    monitor that can re-connect to devices
- Make a heatmap for my lovely Glove80, viewable through a web interface
- Show which key combinations are used the most
- Show key "neighbors" - keys that are pressed in consequence: before or after
  the specific key, but not necessarily in one combination
- Merge keypress data from multiple computers

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

```bash
make build
./tmp/glover track -m explicit -f /dev/tty.usbmodem12301 -f /dev/tty.usbmodem12401 -o keypresses.sqlite -v
```

This will also open a web interface, default location is `localhost:3000`

Interface looks roughly like this:
![preview](img/preview.png)

Alternatively, you can try using auto-detection of the keyboard devices:

```bash
./tmp/glover track -m monitor -o keypresses.sqlite -v
./tmp/glover track -v
```

This will automatically monitor your `/dev/` folder and connect new devices
as they appear.

### Show

In case if you don't need active key tracking, you can only run the web interface

```bash
./tmp/glover show -s keypresses.sqlite -p 8000
```

### Permissions

On some systems, connecting to serial devices might not be available to your
user by default. In order to fix this, you need to add your user to the proper
group. For example:

```bash
> ls -l /dev/ttyACM*
crw-rw---- 1 root uucp 166, 0 Nov 16 13:45 /dev/ttyACM0
crw-rw---- 1 root uucp 166, 1 Nov 16 13:45 /dev/ttyACM1
```

In the output, you can see that these files belong to `uucp` group.
To allow your user to read these files, you need to add your user to that group:

```bash
sudo usermod -a -G uucp $USER
```

## Develop

### Prerequisites

You need following things:

1. `go` for the compiler
2. `npm` for `tailwind` and `prettier`
3. `golangci-lint` if you want to run full CI suite locally.

All 3 of those usually should be installed via your favourite package manager:

```bash
brew install npm go golangci-lint

npm install
```

### Live-reload

For development, it's easy to use [air](https://github.com/air-verse/air).
Make sure to install `npm` to make changes to the css/js related things.

```bash
make run-dev
```

It's possible to just run `air`. In this case, template generation and tailwind
daemon won't run, so only changes to the `.go` files will take effect
(which is sometimes the only needed thing).

### Tests

```bash

# To run specific tests, but all by default
go test ./...

# To run benchmarks
go test -bench -benchtime=10s ./...
# OR, to run only tests
make test
```

To get coverage report, you can use

```bash
make cover
```
