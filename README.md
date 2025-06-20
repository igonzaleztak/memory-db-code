# In-Memory Database

- [In-Memory Database](#in-memory-database)
	- [Overview](#overview)
	- [Design](#design)
	- [Installation and usage](#installation-and-usage)
		- [Set -- POST /api/v1/set](#set----post-apiv1set)
		- [Remove -- DEL /api/v1/test](#remove----del-apiv1test)
		- [Update -- PATCH /api/v1/test](#update----patch-apiv1test)
		- [Push data -- PATCH /api/v1/test/push](#push-data----patch-apiv1testpush)
		- [POP --- PATCH /api/v1/test/pop](#pop-----patch-apiv1testpop)
	- [Optional features](#optional-features)
		- [Data persistence](#data-persistence)
		- [Performance test](#performance-test)
		- [Authentication Module](#authentication-module)


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
	PersistenceEnabled     bool          `mapstructure:"PERSISTENCE_ENABLED"`
	DBPath                 string        `mapstructure:"DB_PATH"` // Optional field that indicates the path where the database is stored
}
```

This config loading is validated using the package [validator](https://github.com/go-playground/validator).

The folder `db` includes the database's logic. The db is very simple, it offers the operations described in the task definition and stores the information within a `map[string]*item`.

```go
// DBClient defines the interface for interacting with an in-memory database.
type DBClient interface {
	// Get retrieves an item by its key.
	Get(key string) (*Item, error)

	// Set stores an item with the specified key and optional options.
	Set(key string, value any, opts ...ItemOptions) error

	// Update modifies an existing item with the specified key and value.
	Update(key string, value any, opts ...ItemOptions) error

	// Remove deletes an item by its key.
	Remove(key string) error

	// Push adds a new item to the memory database with the specified key and value.
	Push(key string, value string, opts ...ItemOptions) (*Item, error)

	// Pop removes and returns the last item from a slice stored at the specified key.
	Pop(key string) (*Item, error)

	// Close releases any resources held by the database client.
	Close()
}

// memoryDB represents an in-memory database that stores items with optional expiration.
type memoryDB struct {
	logger          *slog.Logger     // logger for logging operations
	store           map[string]*Item // in-memory store for items
	cleanupInterval time.Duration    // interval for cleanup routine
	mu              sync.RWMutex     // mutex for preventing race conditions
	stopChan        chan struct{}    // channel to stop the cleanup routine

	// Optional features
	persistenceEnabled bool          // flag to indicate if persistence is enabled
	dbPath             string        // path for persistence storage, if enabled
	logFile            *os.File      // file handle for logging operations, if persistence is enabled
	logEncoder         *json.Encoder // encoder for writing operations to the log file
}
```

`item` struct:

```go
// item represents a single item in the memory database. It would be similar to a row in a traditional database.
type Item struct {
	Value     *StringOrSlice `json:"value"`         // Value can be string or []string
	TTL       time.Time      `json:"ttl,omitempty"` // TTL is optional and will be omitted if not set
	Kind      DataType       `json:"kind"`          // Kind is used internally to determine the data type of the value
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
```

In the previous fragment of code, you can see that the item's value is stored as a custom type called `StringOrSlice`. This type has been created to ensure that the values stored in this struct are either a `string` or a `[]string`. In Go, when you try to unmarshal an array, it is going to be unmarshaled as a `[]interface{}` by default. Therefore, this type implements a custom interface of the `MarshalJSON` and `UnmarshalJSON` interfaces to solve this issue.

```go
// StringOrSlice is a custom type that can hold either a string or a slice of strings.
//
// It implements the json.Unmarshaler and json.Marshaler interfaces to handle JSON serialization and deserialization.
// This ensures that the value is correctly interpreted as either a single string or a slice of strings when unmarshaling from JSON.
type StringOrSlice struct {
	Val any // Value can be string or []string
}

// UnmarshalJSON implements the json.Unmarshaler interface for StringOrSlice.
// It attempts to unmarshal the JSON data into either a string or a slice of strings,
// avoiding it to be unmarshaled into a generic []interface{} type.
func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.Val = str
		return nil
	}

	var slice []string
	if err := json.Unmarshal(data, &slice); err == nil {
		s.Val = slice
		return nil
	}

	return ErrInvalidDataType // Return an error if neither type matches
}

// MarshalJSON implements the json.Marshaler interface for StringOrSlice.
// In this case the marshalJSON method writes the value within the StringOrSlice wrapper and not the wrapper itself.
//
// Example: 
// It will marshal the struct as this:
//
//		{
//			"key": "value",
//			"value": "some string"
//		}
//
// Instead of:
//
//		{
//			"key": "value",
//			"value": {
//				"Val": "some string"
//			}
//		}
func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	switch v := s.Val.(type) {
	case nil:
		return json.Marshal(nil)
	case string:
		return json.Marshal(v)
	case []string:
		return json.Marshal(v)
	default:
		return nil, ErrInvalidDataType
	}
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

**Note:** The previous commands only run the application with the mandatory features.  
Further down in this document, you'll find instructions on how to run the database with optional features enabled.


## API Examples

This section displays some examples on how to interact with the API. Additionally,the API's OpenAPI specification is served in the endpoint [/api/v1/docs/swagger/index.html](http://localhost:8080/api/v1/docs/swagger/index.html).

### Set -- POST /api/v1/set

Input:

```json
{
    "key": "test",
    "value": "Test1",
    "ttl": "1h"
}
```

Response:

```json
{
    "message": "ok"
}
```

### Get  -- GET /api/v1/test

Response:

```json
{
    "key": "test",
    "kind": "string_slice",
    "value": [
        "Test1",
        "test2"
    ],
    "ttl": "2025-06-20T19:40:46.266542+02:00",
    "created_at": "2025-06-20T18:39:59.327178+02:00",
    "updated_at": "2025-06-20T18:40:49.38729+02:00"
}
```

### Remove -- DEL /api/v1/test

Response:

```json
{
    "message": "ok"
}
```

### Update -- PATCH /api/v1/test

Body:

```json
{
    "value": [
        "Test1"
    ],
    "ttl": "1h"
}
```

Response:

```json
{
    "message": "ok"
}
```

### Push data -- PATCH /api/v1/test/push

Body:

```json
{
    "value": "test2",
    "ttl": "24h"
}
```

Response:

```json
{
    "key": "test",
    "kind": "string_slice",
    "value": [
        "Test1",
        "test2"
    ],
    "ttl": "2025-06-20T17:43:50.995712669Z",
    "created_at": "2025-06-20T16:43:47.150490625Z",
    "updated_at": "2025-06-20T16:43:54.357287171Z"
}
```

### POP --- PATCH /api/v1/test/pop

Response: 

```json
{
    "key": "test",
    "kind": "string_slice",
    "value": [
        "Test1",
    ],
    "ttl": "2025-06-20T17:43:50.995712669Z",
    "created_at": "2025-06-20T16:43:47.150490625Z",
    "updated_at": "2025-06-20T16:43:54.357287171Z"
}
```


## Optional features

This section describes the optional features that have been implemented in the project and how they have been implemented.

### Data persistence

Support for data persistence has been added to the in-memory database. This means that when you run the database in persistence mode, the stored data will not be lost after restarts.   To implement this feature, the database logs every operation to a file, reads the file at startup, and replays all the stored commands.


This is how the log file looks like:

```json
{"command":"set","key":"test","time":"2025-06-20T16:54:21.911793+02:00","value":"Test1","ttl":"2025-06-20T17:54:21.911793+02:00","kind":0,"created_at":"2025-06-20T16:54:21.911792+02:00","updated_at":"2025-06-20T16:54:21.911792+02:00"}
{"command":"update","key":"test","time":"2025-06-20T16:54:32.216356+02:00","value":["Test1"],"ttl":"2025-06-20T16:54:42.216356+02:00","kind":1,"created_at":"2025-06-20T16:54:21.911792+02:00","updated_at":"2025-06-20T16:54:32.216355+02:00"}
{"command":"push","key":"test","time":"2025-06-20T16:54:34.434883+02:00","value":"test2","ttl":"0001-01-01T00:00:00Z","kind":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"2025-06-20T16:54:34.434883+02:00"}
{"command":"remove","key":"test","time":"2025-06-20T16:54:37.87889+02:00"}
```

You can easily actiave this feature by setting the following environment variables:

- `PERSISTENCE_ENABLED`: boolean that admits `true` or `false`. If `true` it stores the data persistently.
- `DB_PATH`: Path inside your filesystem where the db will store the data.


By default, persistence is disabled as you can see if you take a look in the main.go file.

```go
	dbOpts := []db.DBOptions{db.WithCleanupInterval(configuration.DefaultCleanupInterval)}
	if configuration.PersistenceEnabled {
		logger.Info("Persistence is enabled, setting up database with persistence options")
		dbOpts = append(dbOpts, db.WithPersistenceEnabled(configuration.DBPath))
	}
	db := db.NewMemoryDB(logger, dbOpts...)
```

The file [persistence.go](internal/db/persistence.go) contains the implementation of this feature.

You can easily run the db in persistence mode by executing the next command:

```bash
task run_persistence
```

Or alternatively, you can uncomment the next global variables in the docker-compose.yaml file.

```yaml
services:
  gomemdb:
    container_name: gomemdb
    build:
      context: .
      dockerfile: Dockerfile
      no_cache: true
    ports:
      - "8080:8080" # Main service port
      - "8081:8081" # Health check port
    environment:
      - API_VERSION="v1.0.0"
      - PORT=8080
      - HEALTH_PORT=8081

    #   # Environment variables for enabling persistence
    #   - PERSISTENCE_ENABLED=true
    #   - DB_PATH=/home/gomemdb

    # volumes:
    #   - .db:/tmp/gomemdb
```

### Performance test

To support this feature, the benchmark file [memory_bench_test.go](internal/db/memory_bench_test.go) has been included. In this file you can see an example of how the memory db could be benchmarked. You can run this benchmark using the command `task benchmark` or alternatively the following command:

```bash
go test -run=^$ -bench=. -benchmem -v ./... -race
```

To get more detailed insights of the Go code I would perform profiling.


### Authentication Module

Unfortunately, I did not have enough time to implement the authentication module, but here is how I would approach it. Assuming the authentication layer is required to protect restricted endpoints, I would proceed as follows:

1. Create an `auth` package inside the `internal` directory. This package would implement the following interface. The `Register` method would be used to register users in the application, hence the inclusion of `adminUser` and `adminPassword` parameters.  Once users log in, they would receive a JWT token signed by the application.
   
	```go
	type AuthManager interface {
		Register(adminUser, adminPassword, username, password string) error
		Login(username, password string) (string, error)
		RefreshToken(token string) (string, error)
		Logout(token string) error
	}
	```

2. Create a user table in the in-memory database to store the username, the hashed password, and the token associated with each user

	```go
	type AuthItem struct {
		UUID uuid.UUID `json:"uuid"`
		Username string `json:"username"`
		Password string `json:"password"` // hashed password.
		Token    string `json:"token"`
	}
	```
	Since a new table is required, it would be added to the in-memory database as a separate map.
	A dedicated `sync.RWMutex` would also be created for this table to avoid locking the main data store when only modifying authentication data.
	
	```go
	type memoryDB struct {
		storeMu    sync.RWMutex
		store      map[string]*Item

		authMu     sync.RWMutex
		authStore  map[uuid.UUID]*AuthItem

		// ...
	}
	```

3. Connect the auth module to the database, then implement the necessary endpoints as described in the `AuthManager` interface so users can authenticate through the platform.
4. Define protected endpoints, and implement a middleware that validates incoming JWTs. The middleware would verify token expiration and signature validity to ensure proper access control.
