//go:build !linux

package display

import "errors"

func New(brightness int, ledcount int) (*Leds, error) {
	return nil, errors.New("LED hardware only available on Linux/Raspberry Pi")
}
