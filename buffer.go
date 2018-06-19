package main

import (
	"fmt"
)

type Buffer struct {
	buffer chan byte
}

func (buf *Buffer) ReadToEnd(p []byte) (n int, err error) {
	for i := range p {
		b, ok := <-buf.buffer
		if !ok {
			return i + 1, fmt.Errorf("closed channel")
		}
		p[i] = b
		if b == '\n' {
			if i > 0 && p[i-1] == '\r' {
				return i + 1, nil
			}
		}
	}
	return len(p), fmt.Errorf("not find end")
}

func (buf *Buffer) Read(p []byte) (n int, err error) {
	for i := range p {
		b, ok := <-buf.buffer
		if !ok {
			return i + 1, fmt.Errorf("closed channel")
		}
		p[i] = b
	}
	return len(p), nil
}

func (buf *Buffer) Write(p []byte) (n int, err error) {
	for _, b := range p {
		buf.buffer <- b
	}
	return len(p), nil
}

func NewBuffer(N int) *Buffer {
	return &Buffer{
		buffer: make(chan byte, N)}
}
