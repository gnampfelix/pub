package pub

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Subscriber", func() {
	var sub Subscriber
	It("should create a new Subscriber", func() {
		sub = NewSubscriber()
		Expect(sub).NotTo(Equal(nil))
	})

	It("should place the message in the channel", func() {
		message := NewMessage("tag")
		message.Write([]byte("abcd"))
		sub.Receive(message)
		simpleSub := sub.(*simpleSubscriber)
		Eventually(simpleSub.channel).Should(Receive())
	})

	It("should return the message from the channel", func() {
		message := NewMessage("tug")
		message.Write([]byte("abcd"))
		sub.Receive(message)
		Eventually(func() string {
			msg := sub.WaitForMessage()
			return msg.Tag()
		}).Should(Equal(message.Tag()))
	})
})
