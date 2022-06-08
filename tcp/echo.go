package main

import (
	"fmt"
	"log"
	"net"
)

func ListenAndServe(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen err: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bind: %s, start listening...", address))
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}

		go Handle(conn)
	}
}

func Handle(conn net.Conn) {

}
