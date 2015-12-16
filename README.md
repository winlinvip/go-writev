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

| NO | OS   | Server      | Client | Syscall            | Bandwidth| Server Command     | 
| ---- | ----- | ----------- | ------ | ------------------ | -------- |  ------------------  |
| 2 | Linux | c++,100%    | c++,90% | write(one-by-one) | 1172MBps | ./server 1985 false true |
| 3 | Linux | c++,86%%    | c++,100%| write(big-buffer) | 2016MBps | ./server 1985 false false |
| 4 | Linux | c++,73%     | c++,99% | writev            | 3576MBps | ./server 1985 true        |
| 5 | Linux | golang,100% | c++,78% | write(one-by-one) | 743MBps  | ./golang 1985 false true |
| 6 | Linux | golang,193% | c++,55% | write(big-buffer) | 1079MBps | ./golang 1985 flase false |

Conclude:

1. The No.4 is the highest performance, to use writev to avoid memory copy and low syscall.
1. The No.5 is the most weak performance, for lots of syscall.
1. The No.6 is the better solution for golang without writev, but waste lots of cpu(golang 193% compare to c++ 73%).

So, I think the writev can improve about 200%~400% performance for streaming server.

Winlin 2015
