// the prototype for go-writev,
// https://github.com/golang/go/issues/13451
package main

import (
	"fmt"
	"flag"
	"os"
	"runtime"
	"net"
	"sync"
	"io"
	"io/ioutil"
)

func main() {
	fmt.Println("the tcp client prototype for go-writev")
	fmt.Println("please read https://github.com/golang/go/issues/13451")

	var server = flag.String("server", "127.0.0.1", "the tcp server to connect to")
	var port = flag.Int("port", 1985, "the tcp port to connect to")
	var nbCpus = flag.Int("cpus", 1, "the cpus to use")
	var nbClients = flag.Int("clients", 1, "the concurrency client to start")

	flag.Usage = func(){
		fmt.Println(fmt.Sprintf("Usage: %v [--server=string] [--port=int] [--cpus=int] [--clients=int] [--header=int] [-h|--help]", os.Args[0]))
		fmt.Println(fmt.Sprintf("	server, the server address. default 127.0.0.1"))
		fmt.Println(fmt.Sprintf("	port, the server port. default 1985"))
		fmt.Println(fmt.Sprintf("	cpus, the cpus to use. default 1"))
		fmt.Println(fmt.Sprintf("	clients, the concurrency client to start. default 1"))
		fmt.Println(fmt.Sprintf("	help, show this help and exit"))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v --port=1985", os.Args[0]))
	}
	flag.Parse()

	runtime.GOMAXPROCS(*nbCpus)
	fmt.Println(fmt.Sprintf("tcp server, server:%v, listen:%v, cpus:%v, clients:%v",
		*server, *port, *nbCpus, *nbClients))

	wg := sync.WaitGroup{}
	wg.Add(*nbClients)
	go func() {
		defer wg.Done()

		var err error
		var addr *net.TCPAddr
		if addr,err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", *server, *port)); err != nil {
			panic(err)
		}
		var c *net.TCPConn
		if c,err = net.DialTCP("tcp", nil, addr); err != nil {
			panic(err)
		}

		for {
			if _,err = io.Copy(ioutil.Discard, c); err != nil {
				fmt.Println("ignore client err", err)
				break
			}
		}
	} ()
	wg.Wait()
	fmt.Println("all clients quit.")
}