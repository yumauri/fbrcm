package log

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
)

// urlRe matches http/https URLs, stopping at whitespace or ANSI reset sequences.
var urlRe = regexp.MustCompile(`"(https?://[^\s\x1b"]+)"|(?m)(https?://[^\s\x1b"]+)`)

// linkify wraps all URLs in a line with OSC 8 hyperlink escape sequences.
// The terminal renders the URL text as-is (preserving existing ANSI color),
// but treats it as a clickable hyperlink.
func linkify(line string) string {
	return urlRe.ReplaceAllStringFunc(line, func(match string) string {
		sub := urlRe.FindStringSubmatch(match)
		url := sub[1] // quoted form
		if url == "" {
			url = sub[2] // unquoted form
		}
		return "\x1b]8;;" + url + "\x1b\\" + url + "\x1b]8;;\x1b\\"
	})
}

// lineSink holds line sink state used by the log package.
type lineSink struct {
	// mu stores mu for lineSink.
	mu sync.Mutex
	// buf stores buf for lineSink.
	buf bytes.Buffer
	// lines stores lines for lineSink.
	lines []string
	// subs stores subs for lineSink.
	subs map[int]chan string
	// nextSub stores next sub for lineSink.
	nextSub int
}

// newLineSink constructs new line sink and returns the resulting value or error.
func newLineSink() *lineSink {
	return &lineSink{
		subs: make(map[int]chan string),
	}
}

// Write writes write for lineSink and returns the resulting state or error.
func (s *lineSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, _ := s.buf.Write(p)
	for {
		line, err := s.buf.ReadString('\n')
		if err != nil {
			break
		}
		s.publishLocked(strings.TrimRight(line, "\r\n"))
	}

	return n, nil
}

// snapshot handles snapshot for lineSink and returns the resulting state or error.
func (s *lineSink) snapshot() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]string, len(s.lines))
	copy(out, s.lines)
	return out
}

// subscribe handles subscribe for lineSink and returns the resulting state or error.
func (s *lineSink) subscribe() (<-chan string, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextSub
	s.nextSub++

	ch := make(chan string, 256)
	s.subs[id] = ch

	return ch, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if sub, ok := s.subs[id]; ok {
			delete(s.subs, id)
			close(sub)
		}
	}
}

// publishLocked handles publish locked for lineSink and returns the resulting state or error.
func (s *lineSink) publishLocked(line string) {
	if line == "" {
		return
	}

	line = linkify(line)
	s.lines = append(s.lines, line)
	for _, ch := range s.subs {
		select {
		case ch <- line:
		default:
		}
	}
}
