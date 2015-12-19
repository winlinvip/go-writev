# go-writev

The prototype program for [golang-writev](https://github.com/golang/go/issues/13451).

## Usage

This program use a tcp server and client to profile the writev, the arch is:
```
+--------+             +---------+
| client +----tcp------+ server  +
+--------+             +---------+
```

To use private TCPConn.Writev from [winlin](https://github.com/winlinvip/go/tree/go1.5-writev):

```
mkdir -p $GOPATH/src/github.com/winlinvip && cd $GOPATH/src/github.com/winlinvip &&
if [[ ! -d go/src ]]; then echo "get writev go."; git clone https://github.com/winlinvip/go; fi &&
cd go && git checkout go1.5-writev && cd $GOROOT/src && sudo mv net net.`date +%s` && 
sudo ln -sf $GOPATH/src/github.com/winlinvip/go/src/net
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
$GOPATH/bin/tcpclient --port=1985
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
