# Logger

Naavik wraps the logging into a generic interface which is used across the project. Devs are advised to use the wrapper through out the project and not the logging libraries directly.

## Log Wrappers
Currently the wrapper is implemented to support two very well known open source logging libraries. 
* Zero logger (Default) [`pkg/logger/zap_logger.go`]
* Zap logger [`pkg/logger/zero_logger.go`]

Logger impl path : `pkg/logger`

## How to use

### Log Level
Log level for the entire project can be set using a exposed command line arg `--log_level`, defaults to info if not set.

### Log key constants
All the generic logging keys are set under `pkg/logger/constants.go` and it's highly recomended to use the contants from here or update if there are any new following the naming conventions.

### Changing the default logger

The default logger is set to use `ZeroLogger` in `pkg/logger/logger.go`. 

```go
func NewLogger() Logger {
	return NewZeroLogger()
}
```

It can be easily switched to any other implementation example ZapLogger by chaging the builder.
```go
func NewLogger() Logger {
	return NewZapLogger()
}
```

### Logging with fields 
#### Context Aware

If you want to log some field on evey log line of a execution path. 
Example: Let's say you want to log handler name on every log line of the handler. The usual way would be to set it in each and every log line with a format specifier which is not ideal.

With context aware logging, the interface gives you below methods to log a field and carries the context fields though the logging context. i.e., as long as we have the same logger used the fields will be logged.
```go
	// Log string value
	WithStr(key string, value string) Logger
	// log bool value
	WithBool(key string, value bool) Logger
	// Log int value
	WithInt(key string, value int) Logger
	// Log any value. Interpreted automatically using reflect
	With(key string, value interface{}) Logger
```
