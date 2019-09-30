.PHONY: gen run get clean

pkg=github.com/ledyba/go-twitter-unmute-unblock

all: .bin/unmute-unblock;

run: all
	.bin/unmute-unblock

gen:
	go generate $(pkg)

.bin/unmute-unblock: gen $(shell find . -type f -name '*.go')
	@mkdir -p .bin
	go build -o $@ $(pkg)

clean:
	rm -rf .bin
	go clean $(pkg)/...
