package ui

import (
	"bytes"
	"io"
	"strings"
)

// MarkdownStripWriter wraps an io.Writer and strips markdown formatting in real-time.
// It buffers incomplete lines to handle markdown patterns that may span chunk boundaries.
type MarkdownStripWriter struct {
	w      io.Writer
	buffer bytes.Buffer
}

// NewMarkdownStripWriter creates a new MarkdownStripWriter that wraps the given writer.
func NewMarkdownStripWriter(w io.Writer) *MarkdownStripWriter {
	return &MarkdownStripWriter{w: w}
}

// Write implements io.Writer. It buffers incoming data and processes complete lines,
// stripping markdown formatting before writing to the underlying writer.
func (m *MarkdownStripWriter) Write(p []byte) (n int, err error) {
	// Buffer incoming data
	m.buffer.Write(p)

	// Process complete lines only (to handle markdown spanning chunks)
	data := m.buffer.String()
	lastNewline := strings.LastIndex(data, "\n")

	if lastNewline >= 0 {
		toProcess := data[:lastNewline+1]
		remaining := data[lastNewline+1:]

		// Strip markdown and write to underlying writer
		cleaned := StripMarkdown(toProcess)
		_, err = m.w.Write([]byte(cleaned))
		if err != nil {
			return len(p), err
		}

		// Keep remaining incomplete line in buffer
		m.buffer.Reset()
		m.buffer.WriteString(remaining)
	}

	return len(p), nil
}

// Flush writes any remaining buffered content to the underlying writer.
// Call this after the command completes to ensure all output is written.
func (m *MarkdownStripWriter) Flush() error {
	if m.buffer.Len() > 0 {
		cleaned := StripMarkdown(m.buffer.String())
		_, err := m.w.Write([]byte(cleaned))
		m.buffer.Reset()
		return err
	}
	return nil
}
