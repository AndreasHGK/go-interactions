package cmd

// Logger is an interface for a logger that can be passed to the command handler.
type Logger interface {
	Debugf(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Debugln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

// NopLogger will discard logged messages.
type NopLogger struct{}

func (NopLogger) Debugf(format string, args ...interface{}) {}
func (NopLogger) Printf(format string, args ...interface{}) {}
func (NopLogger) Warnf(format string, args ...interface{})  {}
func (NopLogger) Errorf(format string, args ...interface{}) {}
func (NopLogger) Fatalf(format string, args ...interface{}) {}

func (NopLogger) Debugln(args ...interface{}) {}
func (NopLogger) Println(args ...interface{}) {}
func (NopLogger) Warnln(args ...interface{})  {}
func (NopLogger) Errorln(args ...interface{}) {}
func (NopLogger) Fatalln(args ...interface{}) {}
