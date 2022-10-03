# sink

Contains interfaces and methods for sinking data to external services. As a general rule, sinks should support all formats of data if possible. 

Each sink must use a select statement to read data from its channel and check if context was cancelled to prevent goroutine leaks (learn more about goroutine leaks [here](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)).

Information for each sink is available in the [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/internal/sink).
