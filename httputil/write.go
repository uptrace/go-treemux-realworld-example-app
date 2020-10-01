package httputil

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/go-realworld-example-app/httputil/httperror"
	"github.com/vmihailenco/treemux"
)

type M map[string]interface{}

func ErrorHandler(w http.ResponseWriter, req treemux.Request, err error) {
	httpErr := httperror.From(err)
	if httpErr.Status >= 500 {
		logrus.WithContext(req.Context()).WithError(err).Error()
	}
	w.WriteHeader(httpErr.Status)
	_ = Write(w, httpErr)
}

func Write(w http.ResponseWriter, res interface{}) error {
	if res == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	switch res := res.(type) {
	case error:
		res = httperror.From(res)
	case string:
		_, _ = io.WriteString(w, res)
		return nil
	case []byte:
		_, _ = w.Write(res)
		return nil
	}

	enc := json.NewEncoder(w)
	return enc.Encode(res)
}
