package debug

import (
	"log"
)

var (
	Level    = (Information | Warning | Error | Fatal)
	_OnFatal *func()
)
var (
	Information = 1 << 0
	Warning     = 1 << 1
	Error       = 1 << 2
	Verbose     = 1 << 3
	Fatal       = 1 << 4
	All         = (Information | Warning | Error | Verbose | Fatal)
)

func SetLevel(i int) {
	Level = i
}

func Log(i int, s ...interface{}) {
	msg := ""
	start := 0

	if len(s) > 1 {
		start = 1
		msg += s[0].(string)
	} else {
		msg += "%s"
	}

	var m string
	switch i {
	case Information:
		if Level&Information > 0 {
			m = "Information: " + msg
		} else {
			return
		}
		break
	case Warning:
		if Level&Warning > 0 {
			m = "Warning: " + msg
		} else {
			return
		}
		break
	case Error:
		if Level&Error > 0 {
			m = "Error: " + msg
		} else {
			return
		}
		break
	case Verbose:
		if Level&Verbose > 0 {
			m = "Verbose: " + msg
		} else {
			return
		}
		break
	case Fatal:
		if Level&Fatal > 0 {
			m = "Fatal: " + msg
		} else {
			if _OnFatal != nil {
				(*_OnFatal)()
			}
			log.Fatalf("program cannot continue")
			return
		}
		break
	}
	if i == Fatal {
		if _OnFatal != nil {
			(*_OnFatal)()
		}
		log.Fatalf(m, s[start:]...)
	} else {
		log.Printf(m, s[start:]...)
	}
}

func Info(v ...interface{}) {
	Log(Information, v...)
}

func Warn(v ...interface{}) {
	Log(Warning, v...)
}

func Err(v ...interface{}) {
	Log(Error, v...)
}

func Ver(v ...interface{}) {
	Log(Verbose, v...)
}

func Fat(v ...interface{}) {
	Log(Fatal, v...)
}

func OnFatal(f func()) {
	_OnFatal = &f
}
