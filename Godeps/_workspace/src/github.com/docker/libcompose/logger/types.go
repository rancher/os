package logger

type Factory interface {
	Create(name string) Logger
}

type Logger interface {
	Out(bytes []byte)
	Err(bytes []byte)
}

type LoggerWrapper struct {
	Err    bool
	Logger Logger
}

func (l *LoggerWrapper) Write(bytes []byte) (int, error) {
	if l.Err {
		l.Logger.Err(bytes)
	} else {
		l.Logger.Out(bytes)
	}
	return len(bytes), nil
}
