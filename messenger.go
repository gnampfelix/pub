package pub

import (
	"errors"
	"net"
	"strconv"
)

//  A Messenger provides methods to simply connect two parties who want
//  to communicate. A Messenger can be used to transport a Publishers messages
//  over the network, but also to stream more complex data.
type Messenger interface {
	TalkTo(remote string) (Connection, error)
	ListenAt(port int, publisher Publisher) error
	StartConversation(key string) (Connection, bool)
	StopListening()
}

func NewMessenger() Messenger {
	return &messenger{connections: make(map[string]Connection)}
}

type messenger struct {
	port          int
	isTalking     bool
	isListening   bool
	stopListening bool
	connections   map[string]Connection //TODO: Mutex!
}

func (m *messenger) TalkTo(remote string) (Connection, error) {
	if m.isListening {
		return nil, errors.New("messenger can't listen and talk at the same time")
	}
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		return nil, err
	}
	m.isTalking = true
	return NewConnection(conn), nil
}

func (m *messenger) ListenAt(port int, publisher Publisher) error {
	if m.isTalking {
		return errors.New("messenger can't listen and talk at the same time")
	}
	m.isListening = true
	m.port = port
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			if m.stopListening {
				conn.Close()
				listener.Close()
				return
			}
			go m.handleNewConnection(conn, publisher)
		}
	}()
	return nil
}

func (m *messenger) handleNewConnection(conn net.Conn, publisher Publisher) {
	newConn := NewConnection(conn)
	connId, err := newConn.ReceiveString()
	if err != nil {
		newConn.Close()
		return
	}
	notification := NewMessage(connId)
	notification.Write([]byte("READY"))
	publisher.Publish(notification)
	m.connections[connId] = newConn
}

func (m *messenger) StartConversation(key string) (Connection, bool) {
	result, ok := m.connections[key]
	if ok {
		delete(m.connections, key)
	}
	return result, ok
}

func (m *messenger) StopListening() {
	if !m.isListening {
		return
	}
	m.stopListening = true
	conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(m.port))
	if err != nil {
		conn.Close()
	}
}
