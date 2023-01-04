package tiny

import (
	"bufio"
	"os"
)

// ConsoleLineHandler specifies a handler function for ConsoleReader.
type ConsoleLineHandler = func(line string)

// ConsoleReader is a Service that actively reads os.Stdin and passes read lines to the underlying handler.
type ConsoleReader struct {
	handler ConsoleLineHandler
}

// NewConsoleReader creates new ConsoleReader.
func NewConsoleReader(handler ConsoleLineHandler) *ConsoleReader {
	return &ConsoleReader{
		handler: handler,
	}
}

// Start implements the interface of Service.
func (c *ConsoleReader) Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		c.handler(line)
	}

	return scanner.Err()
}

// Stop implements the interface of Service.
func (c *ConsoleReader) Stop() {
	_ = os.Stdin.Close()
}
