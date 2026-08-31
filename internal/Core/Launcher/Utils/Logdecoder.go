package utils

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	xmlEventOpenRe  = regexp.MustCompile(`(?i)<log4j:event(\s|>)`)
	xmlEventCloseRe = regexp.MustCompile(`(?i)</log4j:event\s*>`)
	xmlEventSetRe   = regexp.MustCompile(`(?i)</?log4j:eventset\b`)
	xmlPrologRe     = regexp.MustCompile(`^\s*<\?xml\b`)
	xmlTimestampRe  = regexp.MustCompile(`(?is)\btimestamp="(\d+)"`)
	xmlLevelRe      = regexp.MustCompile(`(?is)\blevel="([^"]+)"`)
	xmlThreadRe     = regexp.MustCompile(`(?is)\bthread="([^"]*)"`)
	xmlMessageRe    = regexp.MustCompile(`(?is)<log4j:message\b[^>]*>(.*?)</log4j:message>`)
)

type log4j2Event struct {
	lines []string
}

// Log4j2XMLWriter convierte a texto plano las líneas del output del juego que
// llegan en formato XML de log4j2 (usado por BatMod y otras versiones de
// terceros que declaran logging tipo "log4j2-xml"). Los eventos
// <log4j:Event timestamp=... level=... thread=...><log4j:Message><![CDATA[...]]>
// se emiten como "[HH:mm:ss] [LEVEL] [thread] mensaje"; cualquier otra línea
// se reenvía tal cual. Es seguro para usar como stdout y stderr del proceso.
type Log4j2XMLWriter struct {
	mu      sync.Mutex
	w       io.Writer
	pending []byte
	event   *log4j2Event
}

func NewLog4j2XMLWriter(w io.Writer) *Log4j2XMLWriter {
	return &Log4j2XMLWriter{w: w}
}

func (d *Log4j2XMLWriter) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	total := len(p)
	d.pending = append(d.pending, p...)
	for {
		idx := bytes.IndexByte(d.pending, '\n')
		if idx < 0 {
			break
		}
		line := string(d.pending[:idx])
		d.pending = d.pending[idx+1:]
		if err := d.processLine(strings.TrimSuffix(line, "\r")); err != nil {
			return total, err
		}
	}
	return total, nil
}

// Flush procesa la última línea incompleta y cualquier evento XML abierto.
// Debe llamarse al terminar el proceso, antes de cerrar el archivo de log.
func (d *Log4j2XMLWriter) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.pending) > 0 {
		line := d.pending
		d.pending = nil
		if err := d.processLine(strings.TrimSuffix(string(line), "\r")); err != nil {
			return err
		}
	}
	if d.event != nil {
		return d.flushEvent()
	}
	return nil
}

func (d *Log4j2XMLWriter) processLine(line string) error {
	trimmed := strings.TrimSpace(line)

	switch {
	case d.event != nil:
		d.event.lines = append(d.event.lines, line)
		if xmlEventCloseRe.MatchString(trimmed) {
			return d.flushEvent()
		}
		return nil

	case xmlEventOpenRe.MatchString(trimmed):
		d.event = &log4j2Event{lines: []string{line}}
		if xmlEventCloseRe.MatchString(trimmed) {
			return d.flushEvent()
		}
		return nil

	case xmlEventSetRe.MatchString(trimmed) || xmlPrologRe.MatchString(trimmed):
		// Cabeceras del stream XML de log4j2: no aportan información.
		return nil
	}

	_, err := fmt.Fprintln(d.w, line)
	return err
}

func (d *Log4j2XMLWriter) flushEvent() error {
	ev := d.event
	d.event = nil
	if ev == nil {
		return nil
	}

	block := strings.Join(ev.lines, "\n")

	ts := ""
	if m := xmlTimestampRe.FindStringSubmatch(block); m != nil {
		if ms, err := strconv.ParseInt(m[1], 10, 64); err == nil && ms > 0 {
			ts = time.UnixMilli(ms).Format("15:04:05")
		}
	}
	level := ""
	if m := xmlLevelRe.FindStringSubmatch(block); m != nil {
		level = strings.ToUpper(m[1])
	}
	thread := ""
	if m := xmlThreadRe.FindStringSubmatch(block); m != nil {
		thread = m[1]
	}
	message := ""
	if m := xmlMessageRe.FindStringSubmatch(block); m != nil {
		raw := m[1]
		if strings.Contains(raw, "<![CDATA[") {
			raw = strings.TrimSuffix(strings.TrimPrefix(raw, "<![CDATA["), "]]>")
		}
		message = strings.TrimSpace(html.UnescapeString(raw))
	}
	if message == "" {
		return nil
	}

	var b strings.Builder
	if ts != "" {
		b.WriteString("[" + ts + "] ")
	}
	if level != "" {
		b.WriteString("[" + level + "] ")
	}
	if thread != "" {
		b.WriteString("[" + thread + "] ")
	}
	b.WriteString(message)
	_, err := fmt.Fprintln(d.w, b.String())
	return err
}
