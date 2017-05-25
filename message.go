package pub

import (
	"bytes"
	"io"
)

type Message interface {
	SetTag(tag string)
	Tag() string
	io.ReadWriteCloser
}

func NewMessage(tag string) Message {
	buffer := bytes.NewBuffer(make([]byte, 0))
	return &simpleMessage{tag: tag,
		buffer:   buffer,
		isSealed: false,
	}
}

type simpleMessage struct {
	tag      string
	buffer   *bytes.Buffer
	isSealed bool
}

func (s simpleMessage) Tag() string {
	return s.tag
}

func (s *simpleMessage) SetTag(tag string) {
	s.tag = tag
}

func (s *simpleMessage) Read(p []byte) (n int, err error) {
	n, err = s.buffer.Read(p)
	return n, err
}

func (s *simpleMessage) Write(p []byte) (n int, err error) {
	if s.isSealed {
		return 0, nil
	}
	n, err = s.buffer.Write(p)
	return n, err
}

func (s *simpleMessage) Close() error {
	s.isSealed = true
	return nil
}
