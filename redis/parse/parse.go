package parse

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/LMfrank/godis_project/interface/redis"
	"io"
)

type Payload struct {
	Data redis.Reply
	Err  error
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// 解析报文并返回所有的 reply
func ParseBytes(data []byte) ([]redis.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse0(reader, ch)
	var results []redis.Reply
	for payload := range ch {
		if payload == nil {
			return nil, errors.New("no protocol")
		}
		if payload.Err != nil {
			if payload.Err == io.EOF {
				break
			}
			return nil, payload.Err
		}
		results = append(results, payload.Data)
	}
	return results, nil
}

// 解析报文并返回第一个 payload
func ParseOne(data []byte) (redis.Reply, error) {
	ch := make(chan *Payload)
	reader := bytes.NewReader(data)
	go parse0(reader, ch)
	payload := <-ch
	if payload == nil {
		return nil, errors.New("no protocol")
	}
	return payload.Data, payload.Err
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64
}

func parse0(reader io.Reader, ch chan<- *Payload) {}

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // read normal line
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // read bulk line
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}
