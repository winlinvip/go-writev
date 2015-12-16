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

## Benchmark

| OS | Server | Client | Syscall | Bandwidth | Server Command | 
| --- | ----- | ------ | ------- | --------- |  -----------  |
| Linux | c++,100% | c++,90% | write(one-by-one) | 1172MBps | ./server 1985 false true |
| Linux | c++,86%% | c++,100% | write(big-buffer) | 2016MBps | ./server 1985 false false |
| Linux | c++,73%  | c++,99% | writev        |  3576MBps | ./server 1985 true |


Winlin 2015
