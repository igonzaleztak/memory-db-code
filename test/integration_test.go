package test

import (
	"context"
	"fmt"
	"memorydb/pkg/godb"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	memdbContainer testcontainers.Container
	client         godb.ApiClient
	suite.Suite
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name: "memorydb-integration-test",
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Join(".."),
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp", "8081/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8081/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err, "failed to start in-memory database container")
	s.T().Cleanup(func() {
		err := container.Terminate(ctx)
		s.Require().NoError(err, "failed to terminate in-memory database container")
	})

	s.memdbContainer = container

	// get the port of the started container
	port, err := container.MappedPort(ctx, "8080")
	s.Require().NoError(err, "failed to get container port")

	// start client
	baseURL := "http://127.0.0.1:" + port.Port()
	s.T().Logf("Connecting to in-memory database at %s", baseURL)
	s.client = godb.NewClient(baseURL, "v1")
}

func (s *IntegrationTestSuite) TestGetAndSet() {
	fmt.Println("Running integration test for Get and Set operations")
	// Test setting a value
	setResp, err := s.client.Set("testKey", "testValue", nil)
	s.Require().NoError(err, "failed to set value")
	s.Equal("ok", setResp.Message)

	// Test getting the value
	getResp, err := s.client.Get("testKey")
	s.Require().NoError(err, "failed to get value")
	s.Equal("testValue", getResp.Value)
}

func (s *IntegrationTestSuite) TestRemove() {
	fmt.Println("Running integration test for Remove operation")
	// Set a value first
	_, err := s.client.Set("removeKey", "removeValue", nil)
	s.Require().NoError(err, "failed to set value before remove test")

	// Test removing the value
	removeResp, err := s.client.Remove("removeKey")
	s.Require().NoError(err, "failed to remove value")
	s.Equal("ok", removeResp.Message)
	// Test getting the removed value
	_, err = s.client.Get("removeKey")
	s.Require().Error(err, "expected error when getting removed value")
}

func (s *IntegrationTestSuite) TestUpdate() {
	fmt.Println("Running integration test for Update operation")
	// Set a value first
	_, err := s.client.Set("updateKey", "initialValue", nil)
	s.Require().NoError(err, "failed to set value before update test")

	// Test updating the value
	updateResp, err := s.client.Set("updateKey", "updatedValue", nil)
	s.Require().NoError(err, "failed to update value")
	s.Equal("ok", updateResp.Message)

	// Test getting the updated value
	getResp, err := s.client.Get("updateKey")
	s.Require().NoError(err, "failed to get updated value")
	s.Equal("updatedValue", getResp.Value)
}

func (s *IntegrationTestSuite) TestPush() {
	// Set a value first
	_, err := s.client.Set("pushKey", []string{"initial"}, nil)
	s.Require().NoError(err, "failed to set value before push test")

	// Test pushing a new value
	pushResp, err := s.client.Push("pushKey", "newValue", nil)
	s.Require().NoError(err, "failed to push value")
	s.Require().NotNil(pushResp, "push response should not be nil")

	// Test getting the pushed value
	getResp, err := s.client.Get("pushKey")
	s.Require().NoError(err, "failed to get pushed value")
	s.Equal("string_slice", getResp.Kind, "expected kind to be string_slice")
	s.Len(getResp.Value, 2, "expected two values in the slice after push")
	s.Equal(getResp.Value, []string{"initial", "newValue"}, "expected values to match after push")
}

func (s *IntegrationTestSuite) TestPop() {
	fmt.Println("Running integration test for Pop operation")
	// Set a value first
	_, err := s.client.Set("popKey", []string{"value1", "value2"}, nil)
	s.Require().NoError(err, "failed to set value before pop test")

	// Test popping a value
	popResp, err := s.client.Remove("popKey")
	s.Require().NoError(err, "failed to pop value")
	s.Equal("ok", popResp.Message)

	// Test getting the popped value
	getResp, err := s.client.Get("popKey")
	s.Require().Error(err, "expected error when getting popped value")
	s.Nil(getResp, "expected nil response for popped key")
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
