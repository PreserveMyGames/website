package validate

import (
	"path"
	"strings"
	"unicode"
)

func Slug(slug string) bool {
	if slug == "" || len(slug) > 128 {
		return false
	}
	if strings.Contains(slug, "..") || strings.Contains(slug, "/") || strings.Contains(slug, "\\") {
		return false
	}
	if path.Base(slug) != slug {
		return false
	}
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func StaticPath(p string) bool {
	if p == "" || strings.Contains(p, "..") {
		return false
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return false
	}
	if strings.ContainsRune(p, '\x00') {
		return false
	}
	for _, r := range p {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
