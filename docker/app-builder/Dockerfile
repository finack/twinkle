# Stage 0: Cross-compile rpi_ws281x C library for ARMv6 (hard-float ABI)
FROM debian:bullseye-slim AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    crossbuild-essential-armhf \
    cmake \
    git \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /foundry
RUN git clone --depth 1 https://github.com/jgarff/rpi_ws281x.git \
  && cd rpi_ws281x \
  && mkdir build \
  && cd build \
  && cmake \
     -DCMAKE_C_COMPILER=arm-linux-gnueabihf-gcc \
     -DCMAKE_C_FLAGS="-march=armv6zk -marm -mfpu=vfp" \
     -DCMAKE_SYSTEM_NAME=Linux \
     -DCMAKE_SYSTEM_PROCESSOR=arm \
     -DBUILD_SHARED=OFF \
     -DBUILD_TEST=OFF \
     .. \
  && cmake --build . \
  && make install

# Stage 1: Go builder with ARM hard-float cross-compiler and C library
FROM golang:1.23-bullseye

RUN apt-get update && apt-get install -y --no-install-recommends \
    crossbuild-essential-armhf \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/lib/libws2811.a /usr/local/lib/
COPY --from=build /usr/local/include/ws2811 /usr/local/include/ws2811
