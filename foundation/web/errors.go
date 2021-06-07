package web

import "github.com/pkg/errors"

type FieldError struct{
	Field string `json:"field"`
	Error string `json:"error"`
}

type ErrorResponse struct{
	Error string `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

type Error struct{
	Err error
	Status int
	Fields []FieldError
}


func NewRequestError(err error, status int) error{
	return &Error{err, status, nil}
}

func (err *Error) Error() string{
	return err.Err.Error()
}



type shutdown struct{
	Message string
}

func NewShutdownError(message string) error{
	return &shutdown{message}
}

func (s *shutdown) Error() string{
	return s.Message
}

func IsShutdown(err error) bool{
	if _, ok := errors.Cause(err).(*shutdown); ok{
		return true
	}
	return false
}