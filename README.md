# pub

Pub offers simple publish/subscribe mechanisms.

## Installation
`go get github.com/gnampfelix/pub`. That's it.

## Usage
Create a new Publisher with `pub.New()`. A publisher publishes messages that are tagged with
 ```go
 message := NewMessage("tag")
 message.Write([]byte("payload")
 yourPub.Publish(message)
 ```
 
 A publisher can also receive messages over the network. Therefore, call `yourPub.Gather(8080)`, where 8080 can be any free port.
 Note: Gather will run endless or until you call `yourPub.CancelGathering(8080)`.
 
 To send messages over the network, call `yourPub.AddRemote("urlToRemote:9090")`. Now every published message will be sent to that
 remote as well. If the remote is not listening (= Gathering), the transmission will fail silently.
 
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
