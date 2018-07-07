package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/golang/glog"
)

type Server struct {
	Tubes  map[string]*Tube
	NextID uint64
}

func genJobID() uint64 {
	return atomic.AddUint64(&server.NextID, 1)
}

const (
	STATE_WANTCOMMAND = iota
	STATE_WANTDATA
	STATE_SENDJOB
	STATE_SENDWORD
	STATE_WAIT
	STATE_BITBUCKET
	STATE_CLOSE
)

var (
	flagPersistDir = flag.String("b", "", "binlog persistent dir")
	flagListenPort = flag.String("p", "11300", "port")

	MSG_OUT_OF_MEMORY   = "OUT_OF_MEMORY\r\n"
	MSG_INTERNAL_ERROR  = "INTERNAL_ERROR\r\n"
	MSG_DRAINING        = "DRAINING\r\n"
	MSG_BAD_FORMAT      = "BAD_FORMAT\r\n"
	MSG_UNKNOWN_COMMAND = "UNKNOWN_COMMAND\r\n"
	MSG_EXPECTED_CRLF   = "EXPECTED_CRLF\r\n"
	MSG_JOB_TOO_BIG     = "JOB_TOO_BIG\r\n"
)

var (
	server = &Server{}
)

const (
	LINE_SIZE = 1024
)

func dispatchCommand(conn net.Conn, data []byte, n int) {

}

func GetMinPriJob(client *Client) *Job {
	tube := server.Tubes[client.TubeName]
	for _, job := range tube.Jobs {
		if job.Pri == tube.MinPri {
			return &job
		}
	}

	// no ready job, waiting

	return nil
}

func SplitCmdString(cmd string) []string {
	args := make([]string, 0)
	start := -1
	end := -1
	for i, v := range cmd {
		if v == ' ' {
			if start > end {
				end = i
				args = append(args, cmd[start:end])
				start = end
			}
		} else {
			if start == end {
				start = i
			}
		}
	}
	if start > end {
		args = append(args, cmd[start:])
	}
	return args
}

func doCmd(client *Client) {
	args := SplitCmdString(string(client.Cmd[:len(client.Cmd)-2]))
	switch args[0] {
	// The "put" command is for any process that wants to insert a job into the queue.
	// It comprises a command line followed by the job body:
	// put <pri> <delay> <ttr> <bytes>\r\n
	// <data>\r\n
	case "put":
		if len(args) != 5 {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		pri, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		delay, err := strconv.Atoi(args[2])
		if err != nil {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		ttr, err := strconv.Atoi(args[3])
		if err != nil {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		bodyLen, err := strconv.Atoi(args[4])
		if err != nil {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		bodyLen += 2

		body := make([]byte, bodyLen)
		client.Buffer.Read(body)
		if string(body[bodyLen-2:]) != "\r\n" {
			client.SendMsg(MSG_EXPECTED_CRLF)
			return
		}

		job := &Job{
			ID:      genJobID(),
			Pri:     uint32(pri),
			Delay:   delay,
			TTR:     ttr,
			BodyLen: bodyLen, Body: body}

		client.SaveJob(job)
		client.SendMsg(fmt.Sprintf("INSERTED %d\r\n", job.ID))
	case "delete":
		if len(args) != 2 {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}

		id, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			client.SendMsg(MSG_BAD_FORMAT)
			return
		}
		if err := client.DeleteJob(id); err != nil {
			client.Conn.Write([]byte("NOT_FOUND\r\n"))
		} else {
			client.Conn.Write([]byte("DELETED\r\n"))
		}
	case "reserve", "reserve-with-timeout":
		job := GetMinPriJob(client)
		if job != nil {
			str := fmt.Sprintf("RESERVED %d %d\r\n", job.ID, job.BodyLen-2)
			client.Conn.Write([]byte(str))
			client.Conn.Write(job.Body)
			client.DeleteJob(job.ID)
		}

	default:
		client.SendMsg(MSG_UNKNOWN_COMMAND)

	}

}

func handleConn(conn net.Conn) {
	glog.Infoln("accept from " + conn.RemoteAddr().String())

	client := &Client{
		Conn:     conn,
		Buffer:   NewBuffer(LINE_SIZE),
		TubeName: "default",
		State:    STATE_WANTCOMMAND,
		CmdRead:  0,
		JobChan:  make(chan *Job, 1)}

	data := make([]byte, LINE_SIZE)
	go func() {
		for {
			n, err := conn.Read(data)
			if err != nil {
				glog.Infoln("read error: ", err)
				conn.Close()
				return
			}

			client.Buffer.Write(data[:n])

		}
	}()

	go func() {
		data := make([]byte, 1024)
		for {
			n, err := client.Buffer.ReadToEnd(data)
			if err != nil {
				client.SendMsg(MSG_BAD_FORMAT)
			}

			client.Cmd = string(data[:n])
			doCmd(client)
		}

	}()
}
func main() {
	flag.Parse()
	if *flagPersistDir != "" {

	}

	ln, err := net.Listen("tcp", ":"+*flagListenPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

func init() {
	server.Tubes = make(map[string]*Tube, 0)
	server.Tubes["default"] = &Tube{Jobs: make([]Job, 0), MinPri: math.MaxUint32}
}
