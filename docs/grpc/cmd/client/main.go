package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/finack/twinkle/rpc/metarmap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect to %v: %v", address, err)
	}

	defer conn.Close()

	c := metarmap.NewMetarMapClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p := []*metarmap.Pixel{}
	var n int32 = 0

	for n < 50 {
		fmt.Printf("Building %v", n)
		pixel := new(metarmap.Pixel)
		pixel.Num = proto.Int32(n)
		pixel.Color = "00ff00"
		p = append(p, pixel)
		n++
	}

	r, err := c.SetPixels(ctx, &metarmap.Pixels{Pixels: p})

	if err != nil {
		log.Fatalf("Could not setpixel: %v", err)
	}
	// fmt.Printf("Server returned Num: %v, Color: %v", pixel.Num, pixel.Color)
	fmt.Printf("rcv %v", r)
}
