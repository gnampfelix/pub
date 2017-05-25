package pub

import (
	"net"

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

	Context("gathering over tcp", func() {
		longTag := `Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa.
			Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec,
			pellentesque eu, pretium quis,.`
		var sub1 Subscriber
		var sub2 Subscriber
		remotePub := New()
		remotePub.AddRemote("localhost:9000")
		It("should publish the message to a single subscriber", func() {
			sub1 = pub.Subscribe("taste", newTestSubscriber)
			testSub1 := sub1.(*testSubscriber)
			go pub.Gather(9000)
			input = NewMessage("taste")
			input.Write(inputBytes)
			remotePub.Publish(input)
			Eventually(testSub1.hasReceived).Should(BeTrue())
		})

		It("should not publish the message to the subscriber", func() {
			pub.Unsubscribe("taste", sub1)
			testSub1 := sub1.(*testSubscriber)
			testSub1.received = false
			input = NewMessage("taste")
			input.Write(inputBytes)
			pub.Publish(input)
			Eventually(testSub1.hasReceived).Should(BeFalse())
		})

		It("should publish a message with a longer tag", func() {

			sub1 = pub.Subscribe(longTag, newTestSubscriber)
			testSub1 := sub1.(*testSubscriber)
			input = NewMessage(longTag)
			input.Write(inputBytes)
			remotePub.Publish(input)
			Eventually(testSub1.hasReceived).Should(BeTrue())
		})

		It("should publish a message to two subscribers", func() {
			pub.Unsubscribe(longTag, sub1)
			input = NewMessage("tuste")
			input.Write(inputBytes)
			sub1 = pub.Subscribe("tuste", newTestSubscriber)
			sub2 = remotePub.Subscribe("tuste", newTestSubscriber)

			testSub1 := sub1.(*testSubscriber)
			testSub2 := sub2.(*testSubscriber)
			remotePub.Publish(input)

			Eventually(testSub1.hasReceived).Should(BeTrue())
			Eventually(testSub2.hasReceived).Should(BeTrue())
		})

		It("should cancel gathering", func() {
			pub.CancelGathering(9000)
			conn, err := net.Dial("tcp", "localhost:9000")
			if conn != nil {
				defer conn.Close()
			}
			Expect(err).Should(HaveOccurred())
		})
	})
})
