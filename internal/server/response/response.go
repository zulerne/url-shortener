package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	StatusOK    = "Ok"
	StatusError = "Error"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Ok() Response {
	return Response{Status: StatusOK}
}

func Error(msg string) Response {
	return Response{Status: StatusError, Error: msg}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var msgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("'%s' is required", err.Field()))
		case "url":
			msgs = append(msgs, fmt.Sprintf("'%s' is not a valid url", err.Field()))
		default:
			msgs = append(msgs, fmt.Sprintf("'%s' is invalid", err.Field()))
		}
	}

	return Response{Status: StatusError, Error: strings.Join(msgs, ", ")}
}
