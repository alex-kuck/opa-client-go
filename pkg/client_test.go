package opa

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type opaRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type opaResult struct {
	Valid bool `json:"valid"`
}

func Test_Client_Query(t *testing.T) {
	tests := []struct {
		name    string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler(t))
			defer srv.Close()

			client := NewClient(srv.URL, srv.Client())
			result, err := Query[opaRequest, opaResult](context.Background(), client, "", opaRequest{})

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func (suite *OPASuite) TestOPAClient_Query() {
	tests := []struct {
		name string
		path string
		body opaRequest
		want func(res opaResult, err error)
	}{
		{
			name: "should return error indicating undefined document",
			path: "/foo/bar/baz",
			want: func(res opaResult, err error) {
				suite.Require().ErrorIs(err, ErrDocumentNotFound)
			},
		},
		{
			name: "should return response from OPA",
			path: "/example/allow",
			body: opaRequest{Age: 22},
			want: func(res opaResult, err error) {
				suite.Require().NoError(err)
				suite.Require().True(res.Valid)
			},
		},
	}
	baseUrl := fmt.Sprintf("http://%s:%s", suite.opaContainer.host, suite.opaContainer.port)
	client := NewClient(baseUrl, http.DefaultClient)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			res, err := Query[opaRequest, opaResult](context.TODO(), client, tt.path, tt.body)
			tt.want(res, err)
		})
	}
}

func TestSuite_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}
	suite.Run(t, new(OPASuite))
}

type OPASuite struct {
	suite.Suite
	opaContainer *opaContainer
}

func (suite *OPASuite) SetupSuite() {
	container, err := newOpaContainer()
	suite.Require().NoError(err)

	suite.opaContainer = container
}

func (suite *OPASuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	suite.Require().NoError(suite.opaContainer.container.Terminate(ctx))
}

type opaContainer struct {
	container testcontainers.Container
	host      string
	port      string
}

func newOpaContainer() (*opaContainer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	port := "8181"
	path, err := filepath.Abs("../policies/example.rego")
	if err != nil {
		return nil, err
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "openpolicyagent/opa:edge",
			ExposedPorts: []string{port + "/tcp"},
			AutoRemove:   true,
			WaitingFor:   wait.ForHTTP("/health?bundle=true").WithPort(nat.Port(port + "/tcp")),
			Mounts: testcontainers.Mounts(
				testcontainers.BindMount(path, "/policies/example.rego"),
			),
			Cmd: []string{"run", "--server", "--log-level=debug", "/policies"},
		},
		Started: true,
	})

	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, err
	}

	return &opaContainer{container: container, host: host, port: mappedPort.Port()}, nil
}
