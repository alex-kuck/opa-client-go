package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type opaRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type opaResult struct {
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

func Test_Client_Query(t *testing.T) {
	type args struct {
		path    string
		request opaRequest
	}
	request := opaRequest{
		Name: "Harry Potter",
		Age:  17,
	}
	want := opaResult{
		Message: "You shall not pass",
		Valid:   false,
	}
	path := "/example/access"
	tests := []struct {
		name    string
		args    args
		handler func(t *testing.T) http.Handler
		want    opaResult
		wantErr func(t *testing.T, err error)
	}{
		{
			name: "return error when request fails",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusBadRequest)
				})
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "request to OPA failed")
				assert.ErrorContains(t, err, strconv.FormatInt(http.StatusBadRequest, 10))
			},
		},
		{
			name: "return error when response cannot be parsed",
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("no JSON"))
				})
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "decode OPA response")
			},
		},
		{
			name: "calls OPA with correct request",
			args: args{
				path:    path,
				request: request,
			},
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.Path, "/v1/data"+path), "should prefix path with /v1/data")
					assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
					assert.Equal(t, "application/json", req.Header.Get("Accept"))

					expBytes, _ := json.Marshal(map[string]interface{}{"input": request})
					var expected map[string]interface{}
					_ = json.Unmarshal(expBytes, &expected)

					defer req.Body.Close()
					body, err := io.ReadAll(req.Body)
					var actual map[string]interface{}
					_ = json.Unmarshal(body, &actual)
					assert.Nil(t, err)
					assert.Equal(t, expected, actual)

					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("{}"))
				})
			},
		},
		{
			name: "returns OPA result from response",
			args: args{},
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					content, _ := json.Marshal(want)
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(fmt.Sprintf(`{"result": %s}`, content)))
				})
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler(t))
			defer srv.Close()

			client := NewClient(srv.URL, srv.Client())
			result, err := Query[opaRequest, opaResult](context.Background(), client, tt.args.path, tt.args.request)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
