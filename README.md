# go-writev
The test example for https://github.com/golang/go/issues/13451

## Usage

The c++ version for writev:
```
# server side.
cd cpp && make && ./server 1985 true
# client side.
cd cpp && make && ./client 1985
```

The golang version for writev:
```
# server side.
cd golang && go build . && ./golang 1985 false true
# client side.
cd cpp && make && ./client 1985
```

Winlin 2015
