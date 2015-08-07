package logger

type NullLogger struct {
}

func (n *NullLogger) Out(_ []byte) {
}

func (n *NullLogger) Err(_ []byte) {
}

func (n *NullLogger) Create(_ string) Logger {
	return &NullLogger{}
}
