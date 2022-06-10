package tcp

import (
	"bufio"
	"context"
	"github.com/LMfrank/godis_project/lib/logger"
	"github.com/LMfrank/godis_project/lib/sync/atomic"
	"github.com/LMfrank/godis_project/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type EchoHandler struct {
	activeConn sync.Map       // 保存所有工作状态client的集合
	closing    atomic.Boolean // 关闭状态标识位
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

type EchoClient struct {
	Conn    net.Conn  // tcp 连接
	Waiting wait.Wait // 当服务端开始发送数据时进入waiting，阻止其他goroutine关闭连接
}

// 关闭客户端连接
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 关闭中的 handler 不会处理新连接
	if h.closing.Get() {
		_ = conn.Close()
		return
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}

		// 发送数据前先置位waiting状态，阻止连接被关闭
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// 关闭服务器
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// 逐个关闭连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
