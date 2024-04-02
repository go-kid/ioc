package syslog

var (
	_logger Logger = New(LvInfo)
)

func Trace(v ...any) {
	_logger.Trace(v...)
}
func Tracef(format string, v ...any) {
	_logger.Tracef(format, v...)
}
func Debug(v ...any) {
	_logger.Debug(v...)
}
func Debugf(format string, v ...any) {
	_logger.Debugf(format, v...)
}
func Info(v ...any) {
	_logger.Info(v...)
}
func Infof(format string, v ...any) {
	_logger.Infof(format, v...)
}
func Warn(v ...any) {
	_logger.Warn(v...)
}
func Warnf(format string, v ...any) {
	_logger.Warnf(format, v...)
}
func Error(v ...any) {
	_logger.Error(v...)
}
func Errorf(format string, v ...any) {
	_logger.Errorf(format, v...)
}
func Panic(v ...any) {
	_logger.Panic(v...)
}
func Panicf(format string, v ...any) {
	_logger.Panicf(format, v...)
}
func Fatal(v ...any) {
	_logger.Fatal(v...)
}
func Fatalf(format string, v ...any) {
	_logger.Fatalf(format, v...)
}

func Level(lv Lv) {
	_logger = _logger.Level(lv)
}

func Pref(pref any) Logger {
	return _logger.Pref(pref)
}

func SetLogger(l Logger) {
	_logger = l
}

func GetLogger() Logger {
	return _logger
}
