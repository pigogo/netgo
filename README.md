`netgo` is an fast and small epoll event loop networking framework. It direct use syscalls rather than using the standard Go [net](https://golang.org/pkg/net/) package, and easy to use.



## Features

- Linux only
- Tcp only
- Simple API
- Low memory usage
- Socket option

## Getting Started

### Installing

To start using netgo, install Go and run `go get`:

```sh
$ go get -u github.com/pigogo/netgo
```

This will retrieve the library.