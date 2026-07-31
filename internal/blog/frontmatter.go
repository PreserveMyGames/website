package blog

import (
	"bytes"
	"strings"

	"github.com/PreserveMyGames/website/internal/constants"
)

func splitFrontMatter(data []byte) (frontMatter, body []byte, ok bool) {
	if !bytes.HasPrefix(data, []byte("---")) {
		return nil, data, false
	}

	afterOpen := data[3:]
	if len(afterOpen) == 0 {
		return nil, data, false
	}
	if afterOpen[0] != '\n' && afterOpen[0] != '\r' {
		return nil, data, false
	}

	rest := afterOpen
	rest = bytes.TrimPrefix(rest, []byte("\r\n"))
	rest = bytes.TrimPrefix(rest, []byte("\n"))
	if len(rest) == 0 {
		return nil, data, false
	}

	close := bytes.Index(rest, []byte("\n---"))
	if close < 0 {
		return nil, data, false
	}

	afterClose := rest[close+1:]
	if !bytes.HasPrefix(afterClose, []byte("---")) {
		return nil, data, false
	}

	afterClose = afterClose[3:]
	if len(afterClose) > 0 && afterClose[0] != '\n' && afterClose[0] != '\r' {
		return nil, data, false
	}

	fm := bytes.TrimSpace(rest[:close])
	if len(fm) == 0 {
		return nil, data, false
	}

	body = bytes.TrimSpace(afterClose)
	return fm, body, true
}

func excerpt(desc, body string) string {
	if desc != "" {
		desc = strings.ToValidUTF8(desc, "")
		if desc != "" {
			return desc
		}
	}
	text := strings.TrimSpace(strings.ToValidUTF8(body, ""))
	runes := []rune(text)
	if len(runes) > constants.ExcerptMaxLen {
		return string(runes[:constants.ExcerptMaxLen]) + "..."
	}
	return text
}
