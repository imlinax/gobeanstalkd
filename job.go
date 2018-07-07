package main

type Job struct {
	ID      uint64
	Pri     uint32
	Delay   int
	TTR     int
	BodyLen int
	Body    []byte
}
