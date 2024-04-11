package syslog

type Lv int8

var LvString = map[Lv]string{
	LvTrace: "trace",
	LvDebug: "debug",
	LvInfo:  "info",
	LvWarn:  "warn",
	LvError: "error",
	LvPanic: "panic",
	LvFatal: "fatal",
}

func (l Lv) String() string {
	return LvString[l]
}

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

var String2Lv = map[string]Lv{
	"trace": LvTrace,
	"debug": LvDebug,
	"info":  LvInfo,
	"warn":  LvWarn,
	"error": LvError,
	"panic": LvPanic,
	"fatal": LvFatal,
}

func NewLvFromString(s string) Lv {
	return String2Lv[s]
}
