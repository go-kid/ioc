package strings2

import (
	"strings"
	"unicode/utf8"
)

type SplitConfig struct {
	Sep                 string
	After               bool
	N                   int
	LeftSkipCharacters  []string
	RightSkipCharacters []string
}

type SplitSetting func(s *SplitConfig)

var (
	LeftBlocks                     = []string{"{", "[", "("}
	RightBlocks                    = []string{"}", "]", ")"}
	DefaultSplitBlock SplitSetting = func(s *SplitConfig) {
		s.LeftSkipCharacters = LeftBlocks
		s.RightSkipCharacters = RightBlocks
	}
)

func Split(val, sep string, ops ...SplitSetting) []string {
	var config = &SplitConfig{
		Sep:                 sep,
		After:               false,
		N:                   -1,
		LeftSkipCharacters:  nil,
		RightSkipCharacters: nil,
	}
	for _, op := range ops {
		op(config)
	}
	return SplitWithConfig(val, config)
}

func SplitN(val, sep string, n int, ops ...SplitSetting) []string {
	var config = &SplitConfig{
		Sep:                 sep,
		After:               false,
		N:                   n,
		LeftSkipCharacters:  nil,
		RightSkipCharacters: nil,
	}
	for _, op := range ops {
		op(config)
	}
	return SplitWithConfig(val, config)
}

func SplitWithConfig(s string, config *SplitConfig) []string {
	sep := config.Sep
	n := config.N
	sepSave := 0
	if config.After {
		sepSave = len(sep)
	}
	left := config.LeftSkipCharacters
	right := config.RightSkipCharacters
	if n == 0 {
		return nil
	}
	if sep == "" {
		return explode(s, n)
	}
	if n < 0 {
		n = strings.Count(s, sep) + 1
	}

	if n > len(s)+1 {
		n = len(s) + 1
	}
	a := make([]string, n)
	n--
	i := 0
	for i < n {
		m := Index(s, sep, left, right)
		if m < 0 {
			break
		}
		a[i] = s[:m+sepSave]
		s = s[m+len(sep):]
		i++
	}
	a[i] = s
	return a[:i+1]
}

func IndexSkipBlocks(s, sep string) int {
	return Index(s, sep, LeftBlocks, RightBlocks)
}

func Index(s, sep string, left, right []string) int {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return idx
	}
	//idx := -1
	in := 0
	for i := 0; i < len(s); i++ {
		a := s[i]
		if contains(left, a) {
			in++
			continue
		}
		if contains(right, a) {
			in--
			if in == 0 {
				idx = strings.Index(s[i+1:], sep)
				if idx == -1 {
					return idx
				}
				idx = idx + i + 1
			}
			continue
		}
		if in == 0 && idx <= i {
			if idx != i {
				idx = strings.Index(s[i:], sep) + i
			} else {
				break
			}
		}
	}
	return idx
}

func contains(s []string, b byte) bool {
	for _, b2 := range s {
		if string(b) == b2 {
			return true
		}
	}
	return false
}

func explode(s string, n int) []string {
	l := utf8.RuneCountInString(s)
	if n < 0 || n > l {
		n = l
	}
	a := make([]string, n)
	for i := 0; i < n-1; i++ {
		_, size := utf8.DecodeRuneInString(s)
		a[i] = s[:size]
		s = s[size:]
	}
	if n > 0 {
		a[n-1] = s
	}
	return a
}
