.PHONY: build clean start stop enable disable run setup

BUILD_PLATFORM=v6

build:
	go build -o ./build/twinkle ./cmd/server/main.go

clean:
	rm -Rf ./build

docker-setup:
	echo "Make sure you have the buildx plugin installed (https://github.com/docker/buildx)"
	docker buildx build --platform linux/arm/$(BUILD_PLATFORM) --tag twinkle-builder --file docker/app-builder/Dockerfile .

docker-build:
	docker run --rm -v "$(shell pwd)":/usr/src/twinkle --platform linux/arm/$(BUILD_PLATFORM) \
  -w /usr/src/twinkle twinkle-builder:latest \
	go build -v -o build/"twinkle-arm$(BUILD_PLATFORM)" cmd/server/main.go

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
	cp build/twinkle twinkle/twinkle
	cp config.yaml twinkle/config.yaml
	cp twinkle.service twinkle/twinkle.service
	sudo mv twinkle /home
	sudo ln -s /home/twinkle/twinkle.service /lib/systemd/system/twinkle.service
