package main

import "net"

type Client struct {
	Conn       net.Conn
	Buffer     *Buffer
	State      int
	TubeName   string
	Cmd        string
	CmdRead    int
	CurrentJob *Job
	JobChan    chan *Job
}

func (client *Client) SaveJob(job *Job) {
	tube := server.Tubes[client.TubeName]
	tube.SaveJob(job)
}

func (client *Client) DeleteJob(id uint64) error {
	tube := server.Tubes[client.TubeName]
	return tube.DeleteJob(id)
}

func (client *Client) SendMsg(msg string) (n int, err error) {
	return client.Conn.Write([]byte(msg))
}
