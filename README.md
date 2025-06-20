# In-Memory Database

## Overview

This repository aims to design an in-memory database that must be able to store `string` and `[]string`. The db must support the next operations:

- Get
- Set
- Update
- Remove
- Push for lists
- Pop for lists

The tool must implement the next features:

- Keys with a limited TTL
- Go client API library
- HTTP REST API


Optional features:

- Data persistence
- Perfomance tests
- Authentication

## Design

This section describes how the project has been structured and the line of thought that I have followed to accomplish the definition of the database.

Below, it can be seen how the project has been structured.

```bash
.
├── api # OpenApi specification
├── cmd
│   └── main.go
├── internal
│   ├── apierrors
│   ├── config
│   ├── db
│   ├── enums
│   ├── expiration
│   ├── logger
│   ├── transport
│   └── validator
├── pkg
│   └── godb
├── tests # includes the integration test
├── .dockerignore
├── .gitignore
├── mockery.yaml
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

The `api` folder includes the OpenApi specification of the API. Additionally, the API spec is served in the API in the endpoint [/api/v1/docs/swagger/index.html](http://localhost:8080/api/v1/docs/swagger/index.html).

In `cmd` you can find the `main.go` file that starts the in-memory database. If you take a look to this file, you will see that it does the following tasks:

1. Load the database's configuration from environmental variables.
2. Start the logger.
3. Start the in-memory db.
4. Start the HTTP server.

The `internal` module contains all the packages required to support the database functionality.  
Since the task definition does not specify whether the database should be used as an external package, I assumed that the only package that needs to be exported is the Go API client. As a result, all the core code is stored in this folder, except for the HTTP client, which is located in the `pkg` folder.

The `apierrors` folder contains the definition of well known errors that the API might return.

The folder `config` specifies the code that loads the initial configuration of the application using the package [viper](https://github.com/spf13/viper). Users can setup some variables of the API through environmental variables such as the verbosity level, port in which the API is listening to request, etc.

```go
type Config struct {
	// Common configuration
	Verbose enums.VerboseLevel `mapstructure:"VERBOSE" validate:"required"`

	// API configuration
	ApiVersion string `mapstructure:"API_VERSION" validate:"required"`
	Port       *int   `mapstructure:"PORT" validate:"required"`
	HealthPort *int   `mapstructure:"HEALTH_PORT" validate:"required"`

	// Database configuration
	DefaultTTL             time.Duration `mapstructure:"DEFAULT_TTL" validate:"required"`
	DefaultCleanupInterval time.Duration `mapstructure:"DEFAULT_CLEANUP_INTERVAL" validate:"required"`
}
```

This config loading is validated using the package [validator](https://github.com/go-playground/validator).

The folder `db` includes the database's logic. The db is very simple, it offers the operations described in the task definition and stores the information within a `map[string]*item`.

```go
// DBClient defines the interface for interacting with an in-memory database.
type DBClient interface {
	// Get retrieves an item by its key.
	Get(key string) (*item, error)

	// Set stores an item with the specified key and optional options.
	Set(key string, value any, opts ...ItemOptions) error

	// Update modifies an existing item with the specified key and value.
	Update(key string, value any, opts ...ItemOptions) error

	// Remove deletes an item by its key.
	Remove(key string) error

	// Push adds a new item to the memory database with the specified key and value.
	Push(key string, values []string, opts ...ItemOptions) (*item, error)

	// Pop removes and returns the last item from a slice stored at the specified key.
	Pop(key string) (*item, error)

	// Close releases any resources held by the database client.
	Close()
}


// memoryDB represents an in-memory database that stores items with optional expiration.
type memoryDB struct {
	store           map[string]*item // in-memory store for items
	cleanupInterval time.Duration    // interval for cleanup routine
	mu              sync.RWMutex     // mutex for preventing race conditions
	stopChan        chan struct{}    // channel to stop the cleanup routine
}
```

`item` struct:

```go
// item represents a single item in the memory database. It would be similar to a row in a traditional database.
type item struct {
	Value     any       `json:"value"`         // Value can be string or []string
	TTL       time.Time `json:"ttl,omitempty"` // TTL is optional and will be omitted if not set
	Kind      DataType  `json:"-"`             // Kind is used internally to determine the data type of the value
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```


The folders `enums` and `logger` contain the definition of enums used in the project and the logger used to print information on console. I've implemented the logger [slog](https://go.dev/blog/slog).

The folder `transport` includes the definion of the HTTP API: routes, handlers, schemas, etc. Some interesting stuff regarding this package is that the router has been implemented using [chi](https://github.com/go-chi/chi) and that the healthcheck is done on a different port. The decision to use a separate port for the health check allows monitoring systems to independently verify the service's health without accessing the main API endpoints. This approach ensures the application remains operational while minimizing the risk of overloading the primary API or exposing sensitive information.

The package [Mockery](https://github.com/vektra/mockery) has been used to mock the db for writing the unit tests of the handlers.

Finally, the last package within this module is `validator` which validates the incoming requests to make sure that their format is valid for the API.

The folder `pkg` includes the packages that can be used in other applications. In this case, the only library that has been externalize is the Go HTTP Client that interacts with the db.

```go
// ApiClient defines the interface for interacting with the in-memory database API.
type ApiClient interface {
	// Get retrieves the value associated with a key from the memory database.
	Get(key string) (*ApiResponse, error)

	// Set stores a key-value pair in the memory database.
	Set(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error)

	// Remove deletes a key-value pair from the memory database.
	Remove(key string) (*schemas.OKResponse, error)

	// Update modifies an existing item in the memory database with the specified key and value.
	Update(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error)

	// Push adds a new item to the memory database with the specified key and value.
	Push(key string, value string, ttl *time.Duration) (*ApiResponse, error)

	// Pop removes the last item from a slice stored at the specified key in the memory database.
	Pop(key string) (*ApiResponse, error)
}
```

In the `test` folder, you can find the database integration tests.  To simulate the application as realistically as possible, I decided to use the [testcontainers](https://golang.testcontainers.org/) package, which makes it easy to run the tests against the application's container.


## Installation and usage

The API can be launch using the tasks defined in the `Taskfile.yaml`, so the tool [Task](https://taskfile.dev/installation) must be installed in your computer. The next commands can be used to run tests and launch the API. 

- Run application: `task run`
- Execute tests: `task test`
- Run the integration test: `task integration_test`
- Run the docker compose: `task docker`

Alternatively, if you don't want to install the tool Taskfile, the bash commands used to run the previous orders are shown in the `Taskfile.yaml`.


## Optional features

This section describes the optional features that have been implemented in the project and how they have been implemented.

### Performance test

To support this feature, the benchmark file [memory_bench_test.go](internal/db/memory_bench_test.go) has been included. In this file you can see an example of how the memory db could be benchmarked. You can run this benchmark using the command `task benchmark` or alternatively the following command:

```bash
go test -run=^$ -bench=. -benchmem -v ./... -race
```

To get more detailed insights of the Go code I would perform profiling.