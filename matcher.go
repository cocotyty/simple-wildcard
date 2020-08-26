package simple_wildcard

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

type matcher interface {
	match(buf []byte, next []matcher) (offset int, ok bool)
}

type Raw []byte

func (r Raw) match(str []byte, next []matcher) (offset int, ok bool) {
	if len(str) < len(r) {
		return 0, false
	}
	if bytes.Equal(r, str[:len(r)]) {
		if len(next) != 0 {
			o, ok := next[0].match(str[len(r):], next[1:])
			if !ok {
				return 0, false
			}
			return len(r) + o, true
		}
		offset = len(r)
		if len(str) != offset {
			return 0, false
		}
		return len(r), true
	}
	return 0, false
}

type Range struct {
	From     int
	To       int
	IsRange  bool
	minWidth int
	maxWidth int
}

func (r Range) match(str []byte, next []matcher) (offset int, ok bool) {
	minWidth := r.minWidth
	if len(str) < minWidth {
		return 0, false
	}
	maxWidth := r.maxWidth
	if len(str) < maxWidth {
		maxWidth = len(str)
	}
	for i := minWidth; i <= maxWidth; i++ {
		num, err := strconv.Atoi(string(str[:i]))
		if err != nil {
			continue
		}
		if (r.IsRange && r.To == -1 && num >= r.From) ||
			(!r.IsRange && r.To == -1 && num == r.From) ||
			(num >= r.From && num <= r.To) {
			if len(next) != 0 {
				o, ok := next[0].match(str[i:], next[1:])
				if !ok {
					continue
				}
				return i + o, true
			}
			if len(str) != i {
				continue
			}
			return i, true
		}
	}
	return 0, false
}

type Wildcard struct{}

func (w Wildcard) match(str []byte, next []matcher) (offset int, ok bool) {
	for i := range str {
		if len(next) != 0 {
			o, ok := next[0].match(str[i+1:], next[1:])
			if !ok {
				continue
			}
			return i + 1 + o, true
		}
		return len(str), true
	}
	return 0, false
}

var rangeMatch = regexp.MustCompile(
	`\[([0-9]+)?([:-])?([0-9]*)\]`)

func Match(pattern string, target string) bool {
	if pattern == target {
		return true
	}
	ranges := rangeMatch.FindAllStringSubmatchIndex(pattern, -1)
	var matchers []matcher
	var last = 0
	appendRaw := func(raw string) {
		raws := strings.Split(raw, "*")
		for i, r := range raws {
			if i != 0 {
				matchers = append(matchers, Wildcard{})
			}
			if len(r) == 0 {
				continue
			}
			matchers = append(matchers, Raw(r))
		}
	}
	for _, rg := range ranges {
		if rg[0] != last {
			appendRaw(pattern[last:rg[0]])
		}
		r := Range{-1, -1, true, 1, 1}
		if rg[2] == -1 {
			r.From = 0
		} else {
			if rg[2] == rg[3] {
				r.From = 0
			}
			var err error
			r.From, err = strconv.Atoi(pattern[rg[2]:rg[3]])
			if err != nil {
				continue
			}
			r.minWidth = rg[3] - rg[2]
			r.maxWidth = r.minWidth
		}
		if rg[4] == -1 {
			r.To = -1
			r.IsRange = false
			last = rg[1]
			matchers = append(matchers, r)
			continue
		}

		if rg[6] == -1 || rg[6] == rg[7] {
			r.To = -1
			last = rg[1]
			matchers = append(matchers, r)
			continue
		}
		var err error
		r.To, err = strconv.Atoi(pattern[rg[6]:rg[7]])
		if err != nil {
			continue
		}
		r.maxWidth = rg[7] - rg[6]
		last = rg[1]
		matchers = append(matchers, r)
	}
	appendRaw(pattern[last:])
	offset, ok := matchers[0].match([]byte(target), matchers[1:])
	if !ok {
		return false
	}
	if offset != len(target) {
		return false
	}
	return true
}
