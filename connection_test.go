package pub

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var _ = Describe("Connection", func() {
	var server Messenger
	var client Messenger
	var clientConn Connection
	var serverConn Connection
	var sub Subscriber
	var pub Publisher
	connId := "abcdefg"

	BeforeEach(func() {
		/*
			The OS is not able to free the ports fast enough, therfore
			we generate random ports to use for each test.
		*/
		currentPort := int(rand.Int31n(100)) + 9000
		server = NewMessenger()
		pub = New()
		sub = pub.Subscribe(connId, NewSubscriber)

		server.ListenAt(currentPort, pub)

		client = NewMessenger()
		clientConn, _ = client.TalkTo("127.0.0.1:" + strconv.Itoa(currentPort))
		clientConn.SendString(connId)
		time.Sleep(20 * time.Millisecond)

		serverConn, _ = server.StartConversation(connId)
	})

	AfterEach(func() {
		clientConn.Close()
		serverConn.Close()
		server.StopListening()
	})

	It("should send a message", func() {
		message := NewMessage("tag")
		message.Write([]byte("payload"))
		err := clientConn.SendMessage(message)
		Expect(err).Should(Succeed())
		result, err := serverConn.ReceiveMessage()
		Expect(err).Should(Succeed())
		Expect(result.Tag()).Should(Equal("tag"))
		resultInBytes, err := ioutil.ReadAll(result)
		Expect(err).Should(Succeed())
		Expect(resultInBytes).Should(Equal([]byte("payload\n")))
	})

	It("should not receive a message", func() {
		message := NewMessage("tag")
		message.Write([]byte("payload"))
		err := clientConn.SendMessage(message)
		Expect(err).Should(Succeed())

		_, err = serverConn.ReceiveMessageWithTag("tig")
		Expect(err).Should(HaveOccurred())
	})

	It("should receive a message", func() {
		message := NewMessage("tag")
		message.Write([]byte("payload"))
		err := clientConn.SendMessage(message)
		Expect(err).Should(Succeed())

		result, err := serverConn.ReceiveMessageWithTag("tag")
		Expect(err).Should(Succeed())
		Expect(result.Tag()).Should(Equal("tag"))
		resultInBytes, err := ioutil.ReadAll(result)
		Expect(err).Should(Succeed())
		Expect(resultInBytes).Should(Equal([]byte("payload\n")))
	})

	Context("files", func() {
		var file *os.File
		var result *os.File
		AfterEach(func() {
			file.Close()
			result.Close()
			os.Remove("tmp")
			os.Remove("tmp2")
		})

		It("should send a file", func() {
			fileContent := "abcder lol \n oder \n 13"
			file, err := os.Create("tmp")
			Expect(err).Should(Succeed())
			file.WriteString(fileContent)
			file.Sync()
			file.Close()

			file, err = os.Open("tmp")
			err = clientConn.SendAndCloseFile(file)
			Expect(err).Should(Succeed())

			result, err = serverConn.ReceiveFile("tmp2")
			Expect(err).Should(Succeed())

			resultContent, err := ioutil.ReadAll(result)
			Expect(err).Should(Succeed())
			Expect(string(resultContent)).Should(Equal(fileContent + "\n"))
		})
	})

	It("should stream", func() {
		input := []byte("hallo, Welt!\nhallo, Wald!")
		clientErr := make(chan error, 1)
		serverErr := make(chan error, 1)
		go streamHelper(clientConn, clientErr)
		go streamHelper(serverConn, serverErr)
		Eventually(func() error {
			result := <-clientErr
			return result
		}).Should(Succeed())
		Eventually(func() error {
			result := <-serverErr
			return result
		}).Should(Succeed())

		buf := make([]byte, 10)
		n, err := clientConn.Write(input)
		Expect(err).Should(Succeed())
		Expect(n).Should(Equal(len(input) - 1)) //	\n are not counted!

		n, err = serverConn.Read(buf)
		Expect(err).Should(Succeed())
		Expect(n).Should(Equal(10))
		Expect(buf).Should(Equal(input[:10]))

		buf = make([]byte, 10)
		n, err = serverConn.Read(buf)
		Expect(err).Should(Succeed())
		Expect(n).Should(Equal(10))
		Expect(buf).Should(Equal(input[10:20]))

		buf = make([]byte, 5)
		n, err = serverConn.Read(buf)
		Expect(err).Should(Succeed())
		Expect(n).Should(Equal(5))
		Expect(buf).Should(Equal(input[20:]))

		err = clientConn.StopStream()
		Expect(err).Should(Succeed())

		buf = make([]byte, 5)
		n, err = serverConn.Read(buf)
		Expect(err).Should(HaveOccurred())
	})

})

func streamHelper(conn Connection, result chan error) {
	err := conn.StartStream()
	result <- err
}
