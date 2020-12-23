package grpcserver

import (
	"context"
	"fmt"

	"github.com/finack/twinkle/rpc/metarmap"
)

type Server struct {
	metarmap.UnimplementedMetarMapServer
}

func (s *Server) SetPixel(ctx context.Context, p *metarmap.Pixel) (pixel *metarmap.Pixel, err error) {
	return &metarmap.Pixel{
		Num:   p.Num,
		Color: p.Color,
	}, nil
}

func (s *Server) SetPixels(ctx context.Context, p *metarmap.Pixels) (empty *metarmap.Empty, err error) {
	// fmt.Println("PIXELS: %#v", p)
	// fmt.Println("Length: %v", len(p.Pixels))

	for _, x := range p.Pixels {
		fmt.Printf("Num: %v, Color: %v\n", x.GetNum(), x.GetColor())
	}

	return &metarmap.Empty{}, nil
}
