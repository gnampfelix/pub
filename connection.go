package pub

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/textproto"
	"os"
)

const (
	MESSAGE_PREFIX    = 20
	MESSAGE_TAG       = 201
	MESSAGE_LINE      = 202
	MESSAGE_END       = 203
	STRING            = 301
	FILE_PREFIX       = 40
	FILE_START        = 401
	FILE_END          = 403
	FILE_LINE         = 402
	STREAM_START      = 501
	STREAM_LINE       = 502
	STREAM_END        = 503
	MESSAGE_END_TEXT  = "END MESSAGE"
	FILE_START_TEXT   = "START FILE"
	FILE_END_TEXT     = "END FILE"
	STREAM_START_TEXT = "START STREAM"
	STREAM_END_TEXT   = "END STREAM"
)

//	Connection can be used to transfer items over the network. A Connection can
//	transfer a string, a Message or a file. Furthermore, you can stil stream
//	over a Connection. All items are treated as lines or collections of lines
//	(according to POSIX, all lines must end with \n). For example, if a file is to
//	be send that doesn't end "correctly" with \n, the \n is added on the receiving side.
//	The resulting file will have a \n at the end. An error does not close the connection,
//	although the connection might by out of sync afterwards.
type Connection interface {
	//	Send a message to a remote. The message is closed before sending.
	//	The tag of the message should not contain "\n"-characters, as they
	//	would (currently) break the connection.
	//	Returns an error if the message could not be delivered or if the connections
	//	is set into streaming mode.
	SendMessage(message Message) error

	//	Sends a single string to the remote. The string should not contain "\n"-
	//	characters as they would (currently) break the connection.
	//	Returns an error if the string could not be delivered or if the connection
	//	is set into streaming mode.
	SendString(message string) error

	//	Sends and closes the given file to the remote. Returns an error if the file
	//	could not be read or the connection is set into streaming mode.
	SendAndCloseFile(file *os.File) error

	//	Waits for a message with the given tag. If the received content is not correct
	//	(not a message or a message but with the wrong tag), an error is returned.
	ReceiveMessageWithTag(tag string) (Message, error)

	//	Waits for any message. If the received content is not a message, an error
	//	is returned.
	ReceiveMessage() (Message, error)

	//	Waits for a string. If the received content is not a string, an error is
	//	returned.
	ReceiveString() (string, error)

	//	Waits for a file and saves it with the given filename. The reading position
	//	is set to 0. Returns an error if the received content is not a file or
	//	if the file could not be safed.
	ReceiveFile(filename string) (*os.File, error)

	//	Sets the connection into streaming mode. Waits for the remote to confirm.
	//	Returns an error, if data is received that is not a valid confirmation.
	StartStream() error

	//	Unsets the streaming mode. Does not wait for a remote confirmation.
	StopStream() error

	//	Read and Write can only be used when in streaming mode. Every call to Write
	//	sends the given data at once. All occurences of "\n" are replaced by "\\n".
	//	This is reverted on the reading side. Callers must make sure that no "\\n"
	//	occurs in the input, otherwise it would result in a "\n" in the received
	//	output. A call to Close closes the connection.
	io.ReadWriteCloser
}

//	Converts the given connection into a Connection.
func NewConnection(conn net.Conn) Connection {
	return &connection{conn: textproto.NewConn(conn), readBuffer: bytes.NewBuffer(make([]byte, 0))}
}

type connection struct {
	conn        *textproto.Conn
	readBuffer  *bytes.Buffer
	isStreaming bool
}

func (c *connection) SendMessage(message Message) error {
	if c.isStreaming {
		return errors.New("can't send message when streaming")
	}
	message.Close()
	tag := message.Tag()
	//TODO: Add split for tag if tag contains \n
	err := c.conn.PrintfLine("%d %s", MESSAGE_TAG, tag)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(message)
	for scanner.Scan() {
		err = c.conn.PrintfLine("%d %s", MESSAGE_LINE, scanner.Text())
		if err != nil {
			return err
		}
	}
	err = c.conn.PrintfLine("%d %s", MESSAGE_END, MESSAGE_END_TEXT)
	return err
}

func (c *connection) SendString(message string) error {
	if c.isStreaming {
		return errors.New("can't send string when streaming")
	}
	//TODO: Add split for string if string contains \n
	return c.conn.PrintfLine("%d %s", STRING, message)
}

func (c *connection) SendAndCloseFile(file *os.File) error {
	if c.isStreaming {
		return errors.New("can't send file when streaming")
	}
	err := c.conn.PrintfLine("%d %s", FILE_START, FILE_START_TEXT)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = c.conn.PrintfLine("%d %s", FILE_LINE, scanner.Text())
		if err != nil {
			return err
		}
	}
	file.Close()
	return c.conn.PrintfLine("%d %s", FILE_END, FILE_END_TEXT)
}

func (c *connection) ReceiveMessageWithTag(tag string) (Message, error) {
	if c.isStreaming {
		return nil, errors.New("can't receive message when streaming")
	}
	_, receivedTag, err := c.conn.ReadCodeLine(MESSAGE_TAG)
	if err != nil {
		return nil, err
	}

	if receivedTag != tag {
		return nil, errors.New("the received tag didn't match the expected tag")
	}

	message := NewMessage(tag)
	for {
		currentCode, currentLine, err := c.conn.ReadCodeLine(MESSAGE_PREFIX)
		if err != nil {
			break
		}
		if currentCode == MESSAGE_END {
			return message, nil
		} else if currentCode == MESSAGE_LINE {
			message.Write([]byte(currentLine + "\n"))
		} else {
			return message, errors.New("can't read multiple tag lines")
		}
	}
	return message, err
}

func (c *connection) ReceiveMessage() (Message, error) {
	if c.isStreaming {
		return nil, errors.New("can't receive message when streaming")
	}
	_, tag, err := c.conn.ReadCodeLine(MESSAGE_TAG)
	if err != nil {
		return nil, err
	}
	message := NewMessage(tag)
	for {
		currentCode, currentLine, err := c.conn.ReadCodeLine(MESSAGE_PREFIX)
		if err != nil {
			break
		}
		if currentCode == MESSAGE_END {
			return message, nil
		} else if currentCode == MESSAGE_LINE {
			message.Write([]byte(currentLine + "\n"))
		} else {
			return message, errors.New("can't read multiple tag lines")
		}
	}
	return message, err
}

func (c *connection) ReceiveString() (string, error) {
	if c.isStreaming {
		return "", errors.New("can't receive string when streaming")
	}
	_, result, err := c.conn.ReadCodeLine(STRING)
	return result, err
}

func (c *connection) ReceiveFile(filename string) (*os.File, error) {
	if c.isStreaming {
		return nil, errors.New("can't receive file when streaming")
	}
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	_, _, err = c.conn.ReadCodeLine(FILE_START)
	if err != nil {
		return file, errors.New("no " + FILE_START_TEXT + " found")
	}
	for {
		code, line, err := c.conn.ReadCodeLine(FILE_PREFIX)
		if err != nil {
			return file, err
		}
		if code == FILE_LINE {
			_, err = file.WriteString(line + "\n")
			if err != nil {
				return file, err
			}
		} else if code == FILE_END {
			file.Sync()
			file.Seek(0, 0)
			return file, nil
		} else {
			return file, errors.New("no valid file line received")
		}
	}
}

func (c *connection) StartStream() error {
	if c.isStreaming {
		return nil
	}
	c.isStreaming = true
	err := c.conn.PrintfLine("%d %s", STREAM_START, STREAM_START_TEXT)
	if err != nil {
		return err
	}
	_, _, err = c.conn.ReadCodeLine(STREAM_START)
	return err
}

func (c *connection) StopStream() error {
	if !c.isStreaming {
		return nil
	}
	c.isStreaming = false
	err := c.conn.PrintfLine("%d %s", STREAM_END, STREAM_END_TEXT)
	return err
}

func (c *connection) Write(p []byte) (int, error) {
	if !c.isStreaming {
		return 0, nil
	}

	transformedMessage := bytes.Replace(p, []byte("\n"), []byte("\\n"), -1)
	err := c.conn.PrintfLine("%d %s", STREAM_LINE, string(transformedMessage))
	return len(p), err
}

func (c *connection) Read(p []byte) (int, error) {
	if !c.isStreaming {
		return 0, nil
	}
	max := len(p)

	n, err := c.readBuffer.Read(p)
	if err != nil {
		if err != io.EOF {
			return n, err
		}
	}

	for n < max {
		_, message, err := c.conn.ReadCodeLine(STREAM_LINE)
		if err != nil {
			if convErr, ok := err.(*textproto.Error); ok {
				if convErr.Code == STREAM_END {
					err = io.EOF
					c.isStreaming = false
				}
			}
			return n, err
		}
		stringBytes := bytes.Replace([]byte(message), []byte("\\n"), []byte("\n"), -1)
		i := 0
		for ; n < max && i < len(stringBytes); i++ {
			p[n] = stringBytes[i]
			n++
		}
		if n >= max {
			tmpBuf := stringBytes[i:]
			_, err = c.readBuffer.Write(tmpBuf)
			return n, err
		}

	}
	return n, nil
}

func (c *connection) Close() error {
	return c.conn.Close()
}
