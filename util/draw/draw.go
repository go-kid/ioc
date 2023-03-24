package draw

import (
	"fmt"
	"strings"
)

func Frame(lines []string) string {
	maxLen := 0
	for _, v := range lines {
		if l := len(v); l > maxLen {
			maxLen = l
		}
	}
	line := "+" + strings.Repeat("-", maxLen) + "+"
	var output string
	output += line + "\n"
	for _, s := range lines {
		output += fmt.Sprintf("|%s%s|\n", s, strings.Repeat(" ", maxLen-len(s)))
	}
	return output + line
}
