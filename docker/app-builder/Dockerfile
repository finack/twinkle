# Stage 0 : Build the C library 

FROM buildpack-deps:bookworm-scm AS build

WORKDIR /foundry

RUN apt-get update -y && apt-get install -y \
  build-essential \
  cmake \
  git

RUN git clone https://github.com/jgarff/rpi_ws281x.git \
  && cd rpi_ws281x \ 
  && mkdir build \
  && cd build \ 
  && cmake -D BUILD_SHARED=OFF -D BUILD_TEST=OFF .. \
  && cmake --build . \
  && make install

# Stage 1 : Build a go image with the rpi_ws281x C library and the go wrapper

FROM golang:bookworm
COPY --from=build /usr/local/lib/libws2811.a /usr/local/lib/
COPY --from=build /usr/local/include/ws2811 /usr/local/include/ws2811

RUN go get github.com/rpi-ws281x/rpi-ws281x-go
