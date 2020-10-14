package indexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

type CommandLogger struct {
	m              sync.Mutex
	logs           []*commandLog
	redactedValues []string
}

type commandLog struct {
	command []string
	out     *bytes.Buffer
}

func NewCommandLogger(redactedValues ...string) *CommandLogger {
	return &CommandLogger{
		redactedValues: redactedValues,
	}
}

func (l *CommandLogger) RecordCommand(command []string, stdout, stderr io.Reader) {
	out := &bytes.Buffer{}

	l.m.Lock()
	l.logs = append(l.logs, &commandLog{command: command, out: out})
	l.m.Unlock()

	var m sync.Mutex
	var wg sync.WaitGroup

	readIntoBuf := func(prefix string, r io.Reader) {
		defer wg.Done()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			m.Lock()
			fmt.Fprintf(out, "%s: %s\n", prefix, scanner.Text())
			m.Unlock()
		}
	}

	wg.Add(2)
	go readIntoBuf("stdout", stdout)
	go readIntoBuf("stderr", stderr)
	wg.Wait()
}

func (l *CommandLogger) String() string {
	buf := &bytes.Buffer{}
	for _, log := range l.logs {
		payload := fmt.Sprintf("%s\n%s\n", strings.Join(log.command, " "), log.out)

		for _, v := range l.redactedValues {
			payload = strings.Replace(payload, v, "******", -1)
		}

		buf.WriteString(payload)
	}

	return buf.String()
}
