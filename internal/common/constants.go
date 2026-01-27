package common

const (
	AppBy      = "JSON-from-iCal by github.com/tom-gora"
	AppVersion = "1.0.0"
)

type ExitCode int

const (
	ExitNorm ExitCode = iota
	ExitVer
)

func (ec ExitCode) Int() int {
	return int(ec)
}
