package pub

import (
	"bytes"
	"io"
)

//	A Message represents a streamed byte content connected to a tag.
//	Callers can write to, read from and close it.
type Message interface {
	//	Sets the tag of the Message.
	SetTag(tag string)

	//	Returns the tag of the Message.
	Tag() string

	//	Read from, Write to and Close the Message. After closing the message,
	//	only writing is prohibited, not reading.
	io.ReadWriteCloser
}

//	Returns a Message with the given tag.
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
	sender   string
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
