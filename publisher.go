package pub

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"sync"
)

//  A Publisher can publish messages to different subscribers.
type Publisher interface {
	Gather(port int) error
	Publish(message Message)
	Subscribe(tag string, subCreater func() Subscriber) Subscriber
	AddRemote(remoteAddress string)
	CancelGathering(port int)
	Unsubscribe(tag string, subscriber Subscriber)
}

type publisher struct {
	subscribers map[string][]Subscriber
	sync.RWMutex
	remotes         []string
	cancelGathering map[int]bool
}

func New() Publisher {
	subs := make(map[string][]Subscriber)
	remotes := make([]string, 0)
	cancelGathering := make(map[int]bool)
	result := publisher{
		subscribers:     subs,
		remotes:         remotes,
		cancelGathering: cancelGathering,
	}
	return &result
}

func (p *publisher) Gather(port int) error {
	p.cancelGathering[port] = false
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if p.cancelGathering[port] {
			conn.Close()
			listener.Close()
			return nil
		}
		if err != nil {
			continue
		}
		go p.handleMessage(conn)
	}
}

func (p *publisher) CancelGathering(port int) {
	p.cancelGathering[port] = true
	conn, _ := net.Dial("tcp", ":"+strconv.Itoa(port))
	defer conn.Close()
}

func (p *publisher) Publish(message Message) {
	p.Lock()
	defer p.Unlock()
	message.Close()

	remoteTemplate := NewMessage(message.Tag())
	localTemplate := io.TeeReader(message, remoteTemplate)
	p.publishLocal(message.Tag(), localTemplate)
	p.publishRemote(message.Tag(), remoteTemplate, message.Sender())
}

func (p *publisher) publishRemote(tag string, payload io.Reader, sender string) {
	tagLength := len(tag)
	header := make([]byte, 0)
	for i := tagLength / 255; i > 0; i-- {
		header = append(header, 255)
	}
	header = append(header, byte(tagLength%255))
	header = append(header, []byte(tag)...)
	var writers []io.Writer
	for i := range p.remotes {
		if p.remotes[i] == sender {
			continue
		}
		currentConn, err := net.Dial("tcp", p.remotes[i])
		defer currentConn.Close()
		if err != nil {
			continue
		}
		writers = append(writers, currentConn.(io.Writer))
	}
	writer := io.MultiWriter(writers...)
	writer.Write(header)
	io.Copy(writer, payload)
}

func (p *publisher) publishLocal(tag string, payload io.Reader) {
	if subs, ok := p.subscribers[tag]; ok {
		messages := make([]Message, len(subs))
		writers := make([]io.Writer, len(subs))
		for i := range messages {
			messages[i] = NewMessage(tag)
			writers[i] = messages[i].(io.Writer)
		}
		writer := io.MultiWriter(writers...)
		io.Copy(writer, payload)

		for i := range subs {
			go subs[i].Receive(messages[i])
		}
	}
}

func (p *publisher) Subscribe(tag string, subCreater func() Subscriber) Subscriber {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.subscribers[tag]; !ok {
		p.subscribers[tag] = make([]Subscriber, 0)
	}
	result := subCreater()
	p.subscribers[tag] = append(p.subscribers[tag], result)
	return result
}

func (p *publisher) AddRemote(remoteAddress string) {
	p.Lock()
	defer p.Unlock()
	p.remotes = append(p.remotes, remoteAddress)
}

func (p *publisher) Unsubscribe(tag string, subscriber Subscriber) {
	subs, ok := p.subscribers[tag]
	subPosition := -1
	if ok {
		for i := range subs {
			if subs[i] == subscriber {
				subPosition = i
				break
			}
		}
		if subPosition != -1 {
			p.subscribers[tag] = append(subs[:subPosition], subs[subPosition+1:]...)
		}
	}
}

func (p *publisher) handleMessage(conn net.Conn) {
	defer conn.Close()
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	if buf.Len() == 0 {
		return
	}
	tagLength := 0
	for {
		currentByte := buf.Next(1)
		tagLength += int(currentByte[0])
		if currentByte[0] < 255 {
			break
		}
	}
	tag := string(buf.Next(tagLength))
	payload := buf.Next(buf.Len())
	message := NewMessage(tag)
	message.Write(payload)
	message.SetSender(conn.RemoteAddr().String())
	p.Publish(message)
}
