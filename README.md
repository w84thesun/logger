# logger

This is a simple wrapper on [Zap](https://github.com/uber-go/zap) library.

It provides a simple configuration and a broad set of methods.

Usage example:  

```go
package main

import "github.com/w84thesun/logger" 

func main() {
    log, err := logger.New(logger.DefaultConfig)
    if err != nil {
        panic(err)
    }
    log.Info("some message")
}
```

The result would be:
```json
{
  "level":"info",
  "@timestamp":"2009-11-10T23:00:00Z",
  "message":"some message",
  "service":"awesome-service",
  "namespace":"awesome-namespace"
}
```

Spaces would be trimmed, added here for readability.

To use some custom fields:
```go
package main

import "github.com/w84thesun/logger" 

func main() {
    log, err := logger.New(logger.DefaultConfig)
    if err != nil {
        panic(err)
    }
    log.With(logger.Fields{"some_field": "test", "another_field": 123}).Info("some message")
}
```

Which results in:
```json
{
  "level":"info",
  "@timestamp":"2009-11-10T23:00:00Z",
  "message":"some message",
  "service":"awesome-service",
  "namespace":"awesome-namespace",
  "some_field":"test",
  "another_field":123
}
```