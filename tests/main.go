package main

import (
	"fmt"
	"time"

	"github.com/kr/beanstalk"
)

func main() {
	conn, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
	if err != nil {
		fmt.Println(err)
		return
	}

	id, err := conn.Put([]byte("hello"), 1024, 0, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	id1, body, err := conn.Reserve(time.Second * time.Duration(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	if id != id1 {
		fmt.Println("id not match")
		return
	}

	if string(body) != "hello" {
		fmt.Println("body not match")
		return
	}
}
