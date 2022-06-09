package tcp

import (
	"bufio"
	"context"
	"fmt"
	"github.com/LMfrank/godis_project/godis/interface/tcp"
	"github.com/LMfrank/godis_project/godis/lib/logger"
	"io"
	"log"
	"net"
	"sync"
)

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}

func Handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞直到遇到分隔符 '\n'
		// 遇到分隔符后会返回上次遇到分隔符或连接建立后收到的所有数据, 包括分隔符本身
		// 若在遇到分隔符之前遇到异常, ReadString 会返回已收到的数据和错误信息
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
			} else {
				log.Println(err)
			}
			return
		}
		b := []byte(msg)
		conn.Write(b)
	}
}

func main() {
	ListenAndServe(":8000")
}
