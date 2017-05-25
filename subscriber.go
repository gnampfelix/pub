package pub

type Subscriber interface {
	WaitForMessage() Message
	Receive(message Message)
}

type simpleSubscriber struct {
	channel chan Message
}

func NewSubscriber() Subscriber {
	channel := make(chan Message, 4)
	return &simpleSubscriber{channel}
}

func (s *simpleSubscriber) WaitForMessage() Message {
	result := <-s.channel
	return result
}

func (s *simpleSubscriber) Receive(message Message) {
	s.channel <- message
}
