package httperror

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/vmihailenco/treemux"
)

var (
	errEOF      = BadRequest("eof", "EOF reading HTTP request body")
	ErrNotFound = NotFound("not found")
	ErrInternal = New(http.StatusInternalServerError, "internal", "internal server error")
)

func From(err error) Error {
	switch err {
	case io.EOF:
		return errEOF
	case pg.ErrNoRows:
		return ErrNotFound
	}

	switch err := err.(type) {
	case Error:
		return err
	case *json.SyntaxError:
		return BadRequest("json_syntax", err.Error())
	}

	return ErrInternal
}

func NotFound(msg string, args ...interface{}) Error {
	return New(http.StatusNotFound, "not_found", msg, args...)
}

func BadRequest(code, msg string, args ...interface{}) Error {
	return New(http.StatusBadRequest, code, msg, args...)
}

//------------------------------------------------------------------------------

type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(status int, code, msg string, args ...interface{}) Error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return Error{
		Status:  status,
		Code:    code,
		Message: msg,
	}
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) H() treemux.H {
	return treemux.H{
		"status":  e.Status,
		"code":    e.Code,
		"message": e.Message,
	}
}
