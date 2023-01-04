package tiny

import (
	"bufio"
	"fmt"
	"os"
)

// ConsoleLineHandler specifies a handler function for ConsoleReader.
type ConsoleLineHandler = func(line string)

// ConsoleReader is a Service that actively reads os.Stdin and passes read lines to the underlying handler.
type ConsoleReader struct {
	handler ConsoleLineHandler
	prompt  string
}

// NewConsoleReader creates new ConsoleReader.
func NewConsoleReader(handler ConsoleLineHandler) *ConsoleReader {
	return &ConsoleReader{
		handler: handler,
	}
}

// Prompt sets and enables printing defined prompt before line reading.
func (c *ConsoleReader) Prompt(prompt string) {
	c.prompt = prompt
}

// Start implements the interface of Service.
func (c *ConsoleReader) Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if c.prompt != "" {
			fmt.Print(c.prompt)
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		c.handler(line)
	}

	return scanner.Err()
}

// Stop implements the interface of Service.
func (c *ConsoleReader) Stop() {
	_ = os.Stdin.Close()
}
