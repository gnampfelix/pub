package pub

import (
	"io"
	"sync"
)

//  A Publisher can publish messages to different subscribers.
type Publisher interface {
	Publish(message Message)
	Subscribe(tag string, subCreater func() Subscriber) Subscriber
	Unsubscribe(tag string, subscriber Subscriber)
}

type publisher struct {
	subscribers map[string][]Subscriber
	sync.RWMutex
}

func New() Publisher {
	subs := make(map[string][]Subscriber)
	result := publisher{
		subscribers: subs,
	}
	return &result
}

func (p *publisher) Publish(message Message) {
	p.Lock()
	defer p.Unlock()
	message.Close()

	p.publishLocal(message.Tag(), message)
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
