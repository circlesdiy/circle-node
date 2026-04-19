package handlers

import (
	"bytes"
	"strings"
	"testing"

	"circles.diy/internal/chat"
)

func TestWriteSSEEventProducesValidFraming(t *testing.T) {
	var buf bytes.Buffer
	writeSSEEvent(&buf, chat.Event{
		Name: "msg-chat-123",
		ID:   "message-42",
		Data: "<div>hello</div>",
	})
	out := buf.String()

	if !strings.Contains(out, "event: msg-chat-123\n") {
		t.Fatalf("missing event line: %q", out)
	}
	if !strings.Contains(out, "id: message-42\n") {
		t.Fatalf("missing id line: %q", out)
	}
	if !strings.Contains(out, "data: <div>hello</div>\n") {
		t.Fatalf("missing data line: %q", out)
	}
	if !strings.HasSuffix(out, "\n\n") {
		t.Fatalf("SSE event must terminate with blank line, got %q", out)
	}
}

func TestWriteSSEEventSplitsMultilineData(t *testing.T) {
	var buf bytes.Buffer
	writeSSEEvent(&buf, chat.Event{
		Name: "msg",
		Data: "line1\nline2",
	})
	out := buf.String()
	if !strings.Contains(out, "data: line1\n") || !strings.Contains(out, "data: line2\n") {
		t.Fatalf("multi-line data not split correctly: %q", out)
	}
}

func TestSanitizeSSEDataCollapsesNewlines(t *testing.T) {
	input := "<div>\n  hello\n  world\r\n</div>"
	out := sanitizeSSEData(input)
	if strings.Contains(out, "\n") || strings.Contains(out, "\r") {
		t.Fatalf("expected newlines to be collapsed, got %q", out)
	}
	if !strings.Contains(out, "<div>") || !strings.Contains(out, "hello") {
		t.Fatalf("expected HTML content preserved: %q", out)
	}
}

func TestExtractActiveChatID(t *testing.T) {
	cases := map[string]string{
		"/chat":                 "",
		"/chat/":                "",
		"/chat/abc-123":         "abc-123",
		"/chat/abc-123/":        "abc-123",
		"/chat/abc-123/settings": "abc-123",
	}
	for path, expected := range cases {
		if got := extractActiveChatID(path); got != expected {
			t.Fatalf("extractActiveChatID(%q) = %q, want %q", path, got, expected)
		}
	}
}
