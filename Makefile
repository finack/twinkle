.PHONY: build clean start stop enable disable run setup

build:
	go build -o ./build/twinkle ./cmd/server/main.go

clean:
	rm -Rf ./build

start:
	sudo service twinkle start

stop:
	sudo service twinkle stop

status:
	sudo service twinkle status

run: build
	sudo build/twinkle config.yaml

enable:
	sudo systemctl enable twinkle

disable:
	sudo systemctl disable twinkle

# Not tested, more documentation
setup: build
	mkdir twinkle
	ln -s build/twinkle twinkle/twinkle
	ln -s config.yaml twinkle/config.yaml
	ln -s twinkle.service twinkle/twinkle.service
	sudo mv twinkle /home
	ln -s /home/twinkle/twinkle.service /lib/systemd/system/twinkle.service
