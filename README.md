# go-hclog

[![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[godocs]: https://godoc.org/github.com/hashicorp/go-hclog

`go-hclog` is a package for Go that provides a simple key/value logging
interface for use in development and production environments.

It provides logging levels that provide decreased output based upon the
desired amount of output, unlike the standard library `log` package.

It does not provide `Printf` style logging, only key/value logging that is
exposed as arguments to the logging functions for simplicity.

It provides a human readable output mode for use in development as well as
JSON output mode for production.

## Installation and Docs

Install using `go get github.com/hashicorp/go-hclog`.

Full documentation is available at
http://godoc.org/github.com/hashicorp/go-hclog

## Usage

**Create a new logger**

```go
logger := log.New(&LoggerOptions{Name: "prog", Level: log.INFO})
```

**Use the global logger**
```go
log.Default()
```

or

```go
log.Default()
```

**Emit an Info level message with 2 key/value pairs**

```go
host := current()
if err := op(); err != nil {
  logger.Info("host encountered an error", "host", host, "error", err)
}
```

**Create a new Logger for a major subsystem**

```go
sub := logger.Named("transport")
```

Logs emitted by `sub` will contain the name field of `prog.transport`,
where `prog` is the name that `logger` currently has.

**Create a new Logger with fixed key/value pairs**

```go
sub := logger.With("request", requestId)
```

Logs emitted by `sub` will always contain the key `requset` with the given
value. This allows sub Loggers to be context specific without having to
thread that into all the callers.

**Use this with code that uses the standard library logger**

```go
import golog 'log'

func processData(l *golog.Log)

processData(
  log.Default().StandardLogger(
    &StandardLoggerOptions{
      InfereLevels: true,
    },
  ),
)
```

This allows this logger to be used in places where the standard library
`*log.Log` is expected. Additionally, it can infer the log levels from
commonly used prefixes used. See the docs for more information.

