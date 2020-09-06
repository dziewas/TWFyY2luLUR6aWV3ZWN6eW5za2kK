package main

import "io"

type Converter struct {
	buf  []byte
	conv func(byte) byte
}

func NewConverter(conv func(byte) byte) *Converter {
	return &Converter{conv: conv}
}

func (c *Converter) Read(p []byte) (int, error) {
	if len(c.buf) == 0 {
		return 0, io.EOF
	}

	n := 0

	if len(p) > len(c.buf) {
		n = copy(p, c.buf)
	} else {
		n = copy(p, c.buf[0:len(p)])
	}

	c.buf = c.buf[n:]
	return n, nil
}

func (c *Converter) Write(p []byte) (int, error) {
	for _, b := range p {
		c.buf = append(c.buf, c.conv(b))
	}

	return len(p), nil
}
