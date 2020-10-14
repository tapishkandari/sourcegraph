package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

type Logger struct {
	m              sync.Mutex
	redactedValues []string
	logs           []*log
}

type log struct {
	command []string
	out     *bytes.Buffer
}

func NewLogger(redactedValues ...string) *Logger {
	return &Logger{
		redactedValues: redactedValues,
	}
}

func (l *Logger) RecordCommand(command []string, stdout, stderr io.Reader) {
	out := &bytes.Buffer{}

	l.m.Lock()
	l.logs = append(l.logs, &log{command: command, out: out})
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

func (l *Logger) String() string {
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
