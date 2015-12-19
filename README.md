# go-writev

The prototype program for [golang-writev](https://github.com/golang/go/issues/13451).

## Usage

This program use a tcp server and client to profile the writev, the arch is:
```
+--------+             +---------+
| client +----tcp------+ server  +
+--------+             +---------+
```

To start the tcp server, use write to send tcp packets:

```
go get -d github.com/winlinvip/go-writev/tcpserver && 
cd $GOPATH/src/github.com/winlinvip/go-writev/tcpserver && go build -a . &&
./tcpserver --port=1985 --writev=false
```

Or use writev to send:

```
go get -d github.com/winlinvip/go-writev/tcpserver && 
cd $GOPATH/src/github.com/winlinvip/go-writev/tcpserver && go build -a . &&
./tcpserver --port=1985 --writev=true
```

Then, please start the client to recv tcp packets:

```
go get github.com/winlinvip/go-writev/tcpclient && 
$GOPATH/bin/tcpserver --port=1985
```

Remarks:

1. Profile: Both server and client will write cpu(cpu.prof) and memory(mem.prof) profile, 
user can use `go tool pprof $GOPATH/bin/tcpserver cpu.prof` to profile it.
1. GO sdk: Please use private implementation for [writev](https://github.com/winlinvip/go/pull/1#issuecomment-165943222).
1. Fast build: User can rebuild the golang sdk for this program by `cd $GOPATH/src/github.com/winlinvip/go-writev/tcpserver && go build -a .`

## Benchmark

Coming soon...

## Conclude

Coming soon...

Winlin 2015
