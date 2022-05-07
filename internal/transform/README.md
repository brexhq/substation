# transform

Contains interfaces and methods for transforming data as it moves from a source to a sink.

Each transform must select from both the data and kill channels to prevent goroutine leaks (learn more about goroutine leaks [here](https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html)).

Information for each transform is available in the [GoDoc](https://pkg.go.dev/github.com/brexhq/substation/internal/transform).
