# hclogslog

This is a go module that allows [hclog](https://github.com/hashicorp/go-hclog) to be used with
[slog](https://pkg.go.dev/log/slog), a structured logging module that is
included in Go 1.21 standard library.

## Usage

To create a slog logger from an existing `hclog.Logger` value:

```

var existing hclog.Logger

...

log := slog.New(hclogslog.Adapt(existing))

```


### Limitations

Today, slog provides Time and PC values to the handlers. Currently these are ignored by the adapter because hclog
doesn't provide a way for these to be overriden per-log message. If a future version of hclog
adds the ability to process them per log, this adapter will be updated.