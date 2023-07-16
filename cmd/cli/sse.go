package main

import (
	"bufio"
	"io"
	"strings"
)

type message struct {
	event string
	data  string
}

func sse(r io.Reader, cb func(e message, err error, done bool)) {
	scanner := bufio.NewScanner(r)

	event := ""
	data := ""
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "event":
			event = strings.TrimSpace(parts[1])
		case "data":
			data = strings.TrimSpace(strings.Join(parts[1:], ":"))
			cb(message{event, data}, nil, false)
		}
	}

	if err := scanner.Err(); err != nil {
		cb(message{}, err, false)
	}

	cb(message{}, nil, true)
}
