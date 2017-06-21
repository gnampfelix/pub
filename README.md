# pub

Pub offers simple publish/subscribe mechanisms.

__Note:__ This package is still in Development and should therefore not be considered stable!

## Installation
`go get github.com/gnampfelix/pub`. That's it.

## Usage
Create a new Publisher with `pub.New()`. A publisher publishes messages that are tagged with
 ```go
 message := NewMessage("tag")
 message.Write([]byte("payload")
 yourPub.Publish(message)
 ```

 After a message is published, it gets closed so that no more writing to that message is possible.

 To subscribe to a specific tag, call:
 ```go
 sub := yourPub.Subscribe("tag", pub.NewSubscriber)
 ```
 The second argument can be any func of type
 ```go
 func() Subscriber
 ```
 To read incoming messages, call
 ```go
 incomingMessage := sub.WaitForMessage()
 ```
 This will return a new message or block. The Subscriber that is included in pub buffers up to 4 messages.

 To unsubscribe from a tag, call
 ```go
 yourPub.Unsubscribe("tag", sub)
 ```

 A full godoc documentation will be added soon.
