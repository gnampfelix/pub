package pub

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testSubscriber struct {
	received bool
}

func (t *testSubscriber) WaitForMessage() Message {
	return nil
}

func (t *testSubscriber) Receive(message Message) {
	t.received = true
}

func newTestSubscriber() Subscriber {
	return &testSubscriber{}
}

func (t *testSubscriber) hasReceived() bool {
	return t.received
}

var _ = Describe("Publisher", func() {
	var pub Publisher
	var sub Subscriber
	var input Message

	var inputBytes = []byte("This is a message regarding the tag \"test\"")
	Context("single subscriber", func() {
		var testSub *testSubscriber
		It("should create a new Publisher", func() {
			pub = New()
			Expect(pub).NotTo(Equal(nil))
		})

		It("should return a Subscriber", func() {
			sub = pub.Subscribe("test", newTestSubscriber)
			testSub = sub.(*testSubscriber)
			Expect(sub).NotTo(Equal(nil))
		})

		It("should publish the message to the sub", func() {
			input = NewMessage("test")
			input.Write(inputBytes)
			pub.Publish(input)
			Eventually(testSub.hasReceived).Should(BeTrue())
		})

		It("should not publish the message to the sub", func() {
			testSub.received = false
			input = NewMessage("toast")
			input.Write(inputBytes)
			pub.Publish(input)
			Consistently(testSub.hasReceived).Should(BeFalse())
		})
	})
})
