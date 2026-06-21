.PHONY: build clean start stop enable disable run setup test lint deps-upgrade deploy

BUILD_PLATFORM=v6
DEPLOY_HOST=peter@192.168.7.162
DEPLOY_PATH=/home/twinkle

test:
	go test ./internal/...

lint:
	go vet ./...
	gofmt -l .
	staticcheck ./...

deps-upgrade:
	go get -u ./...
	go mod tidy

build:
	go build -o ./build/twinkle ./cmd/server/main.go

clean:
	rm -Rf ./build

docker-setup:
	docker build --tag twinkle-builder --file docker/app-builder/Dockerfile .

docker-build:
	mkdir -p build
	docker run --rm -v "$(shell pwd)":/usr/src/twinkle \
	  -w /usr/src/twinkle twinkle-builder:latest \
	  env CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc \
	      CGO_CFLAGS="-march=armv6zk -marm -mfpu=vfp" \
	      GOARCH=arm GOARM=$(BUILD_PLATFORM) GOOS=linux \
	  go build -v -o build/twinkle-arm cmd/server/main.go

deploy:
	rsync -avz --exclude='.git' --exclude='build/' --exclude='vendor/' . $(DEPLOY_HOST):/home/peter/twinkle-src/
	ssh $(DEPLOY_HOST) "cd /home/peter/twinkle-src && /usr/local/go/bin/go build -o $(DEPLOY_PATH)/twinkle.new ./cmd/server/main.go"
	ssh $(DEPLOY_HOST) "sudo systemctl stop twinkle && mv $(DEPLOY_PATH)/twinkle.new $(DEPLOY_PATH)/twinkle && sudo systemctl start twinkle"

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
