package testbed

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/uptrace/go-realworld-example-app/org"
	"github.com/uptrace/go-realworld-example-app/rwe"

	. "github.com/onsi/gomega"
)

func setToken(req *http.Request, userID uint64) {
	if userID == 0 {
		return
	}

	token, err := org.CreateUserToken(userID, time.Hour)
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Authorization", "Token "+token)
}

func ParseJSON(resp *httptest.ResponseRecorder, code int) map[string]interface{} {
	res := make(map[string]interface{})
	err := json.Unmarshal(resp.Body.Bytes(), &res)
	Expect(err).NotTo(HaveOccurred())

	Expect(resp.Code).To(Equal(code))

	return res
}

func serve(req *http.Request) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	rwe.Router.ServeHTTP(resp, req)
	return resp
}

func Get(url string) *httptest.ResponseRecorder {
	return GetWithToken(url, 0)
}

func GetWithToken(url string, userID uint64) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	setToken(req, userID)
	return serve(req)
}

func Post(url, data string) *httptest.ResponseRecorder {
	return PostWithToken(url, data, 0)
}

func PostWithToken(url, data string, userID uint64) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "application/json")
	setToken(req, userID)
	return serve(req)
}

func PutWithToken(url, data string, userID uint64) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PUT", url, bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "application/json")
	setToken(req, userID)
	return serve(req)
}

func DeleteWithToken(url string, userID uint64) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")
	setToken(req, userID)
	return serve(req)
}
