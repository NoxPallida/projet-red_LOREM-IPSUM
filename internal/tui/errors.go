package tui

import (
	"runtime"
	"strconv"
)

type GameError struct {
	File      string
	Line      int
	Operation string
	Err       error
}

func NewError(operation string, Err error) *GameError {
	// runtime.Caller permet de recuperer le nom du fichier qui appel la fonction + autre args
	_, file, line, _ := runtime.Caller(1)
	return &GameError{
		File:      file,
		Line:      line,
		Operation: operation,
		Err:       Err,
	}
}

func ErrorText(e *GameError) string {
	strError := ""
	if e.Err == nil {
		strError = e.File + " : " + e.Operation
	} else {
		strError = e.File + ":" + strconv.Itoa(e.Line) + e.Operation + " -> " + e.Err.Error()
	}
	return strError
}

func (e *GameError) Error() string { return ErrorText(e) }
