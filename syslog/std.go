package syslog

type Lv int8

const (
	LvTrace Lv = iota + 1
	LvDebug
	LvInfo
	LvWarn
	LvError
	LvPanic
	LvFatal
)

var lvPrefix = map[Lv]string{
	LvTrace: "[TRACE]",
	LvDebug: "[DEBUG]",
	LvInfo:  "[ INFO]",
	LvWarn:  "[ WARN]",
	LvError: "[ERROR]",
	LvPanic: "[PANIC]",
	LvFatal: "[FATAL]",
}

var lvColor = map[Lv]color{
	LvTrace: clStd,
	LvDebug: clBlue,
	LvInfo:  clGreen,
	LvWarn:  clYellow,
	LvError: clRed,
	LvPanic: clRedB,
	LvFatal: clRedB,
}
