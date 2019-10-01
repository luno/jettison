package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-stack/stack"
	"github.com/luno/jettison/models"
)

func NewError(msg string) models.Error {
	return models.Error{
		Message: msg,
		Source:  fmt.Sprintf("%+v", stack.Caller(2)),
		Code:    msg,
	}
}

func NewHop() models.Hop {
	return models.Hop{
		Binary: filepath.Base(os.Args[0]),
	}
}
