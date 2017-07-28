package cache

type NilLogger struct {
}

func NewNilLogger() *NilLogger {
	return &NilLogger{}
}

func (this *NilLogger) IsDebugEnabled() bool {
	return false
}

func (this *NilLogger) Debugf(message string, args ...interface{}) {
}

func (this *NilLogger) Debug(...interface{}) {
}

func (this *NilLogger) Warningf(message string, args ...interface{}) {
}

func (this *NilLogger) Warning(...interface{}) {
}

func (this *NilLogger) Errorf(message string, args ...interface{}) {
}

func (this *NilLogger) Criticalf(message string, args ...interface{}) {
}

func (this *NilLogger) Printf(format string, v ...interface{}) {
}

func (this *NilLogger) Critical(...interface{}) {
}
