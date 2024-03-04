package syslog

type color int8

const (
	clStd color = iota + 1
	clRed
	clGreen
	clYellow
	clBlue
	clRedB
	clGreenB
	clYellowB
	clBlueB
)

var clCode = map[color]string{
	clStd:     "",
	clRed:     "\033[31m",
	clGreen:   "\033[32m",
	clYellow:  "\033[33m",
	clBlue:    "\033[34m",
	clRedB:    "\033[31;1m",
	clGreenB:  "\033[32;1m",
	clYellowB: "\033[33;1m",
	clBlueB:   "\033[34;1m",
}

const reset = "\033[0m"
