// the prototype for go-writev,
// https://github.com/golang/go/issues/13451
package main

import (
	"fmt"
	"os"
	"runtime"
	"flag"
	"net"
	"os/signal"
	"runtime/pprof"
)

func main() {
	fmt.Println("the tcp server prototype for go-writev")
	fmt.Println("please read https://github.com/golang/go/issues/13451")

	var port = flag.Int("port", 1985, "the tcp server listen at")
	var useWritev = flag.Bool("writev", false, "whether use writev")
	var nbCpus = flag.Int("cpus", 1, "the cpus to use")
	var cpuProfile = flag.String("cpup", "cpu.prof", "the cpu profile file")
	var memProfile = flag.String("memp", "mem.prof", "the memory profile file")
	var groupSize = flag.Int("group", 512, "the size of group, which contains N*(header, payload)")
	var headerSize = flag.Int("header", 12, "the size of header")
	var payloadSize = flag.Int("payload", 1024, "the size of payload")

	flag.Usage = func(){
		fmt.Println(fmt.Sprintf("Usage: %v [--port=int] [--writev=bool] [--cpus=int] [--cpup=string] [--memp=string] [--group=int] [--header=int] [--payload=int] [-h|--help]", os.Args[0]))
		fmt.Println(fmt.Sprintf("	port, the listen port. default 1985"))
		fmt.Println(fmt.Sprintf("	writev, whether use writev to send. default false(use write)"))
		fmt.Println(fmt.Sprintf("	cpus, the cpus to use. default 1"))
		fmt.Println(fmt.Sprintf("	cpup, the cpu profile file. default cpu.prof"))
		fmt.Println(fmt.Sprintf("	memp, the memory profile file. default mem.prof"))
		fmt.Println(fmt.Sprintf("	group, the size of group, which contains N*(header, payload). default 512"))
		fmt.Println(fmt.Sprintf("	header, the size of header. default 12"))
		fmt.Println(fmt.Sprintf("	payload, the size of payload. default 1024"))
		fmt.Println(fmt.Sprintf("	help, show this help and exit"))
		fmt.Println(fmt.Sprintf("Remark:"))
		fmt.Println(fmt.Sprintf("	use SIGINT or ctrl+c to interrupt program and collect cpu/mem profile data."))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v --writev=false", os.Args[0]))
	}
	flag.Parse()

	runtime.GOMAXPROCS(*nbCpus)
	fmt.Println(fmt.Sprintf("tcp server, writev:%v, listen:%v, cpus:%v, cpuProfile:%v, memProfile:%v, group:%v, header:%v, payload:%v",
		*useWritev, *port, *nbCpus, *cpuProfile, *memProfile, *groupSize, *headerSize, *payloadSize))
	fmt.Println("use SIGINT or ctrl+c to interrupt program and collect cpu/mem profile data.")

	// always start profile.
	if w,err := os.Create(*cpuProfile); err != nil {
		panic(err)
	} else if err := pprof.StartCPUProfile(w); err != nil {
		panic(err)
	}
	// use signal SIGINT to profile and quit.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func(){
		<-c
		fmt.Println("To profile program:")
		pprof.StopCPUProfile()
		fmt.Println("go tool pprof", os.Args[0], *cpuProfile)
		if w,err := os.Create(*memProfile); err != nil {
			panic(err)
		} else if err := pprof.Lookup("heap").WriteTo(w, 0); err != nil {
			panic(err)
		}
		fmt.Println("go tool pprof", os.Args[0], *memProfile)
		os.Exit(0)
	}()

	var err error
	var l *net.TCPListener
	if l, err = srs_listen(*port); err != nil {
		panic(err)
	}

	// create group of tcp packets to send.
	// each group is a serial of header and payload,
	// 		group = h0, p0, h1, p1, ..., hN, pN
	group := make([][]byte, *groupSize * 2)
	for i := 0; i < len(group); i += 2 {
		// create header.
		header := make([]byte, *headerSize)
		for j,_ := range header {
			header[j] = 0xf0 + byte(i) + byte(j)
		}
		// create payload
		payload := make([]byte, *payloadSize)
		for j,_ := range payload {
			payload[j] = 0x0f + byte(i) + byte(j)
		}
		// the group contains N*(header, payload)
		// h0, p0, h1, p1, ..., hN, pN
		group[i] = header
		group[i+1]=payload
	}

	for {
		var c *net.TCPConn
		if c,err = l.AcceptTCP(); err != nil {
			panic(err)
		}
		if err = srs_serve(c, *useWritev, group); err != nil {
			fmt.Println("ignore serve client err", err)
		}
	}
}

func srs_listen(port int) (l *net.TCPListener, err error) {
	var addr *net.TCPAddr
	if addr,err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", port)); err != nil {
		return
	}
	if l,err = net.ListenTCP("tcp", addr); err != nil {
		return
	}
	fmt.Println("listen at", fmt.Sprintf("tcp://%v", addr))
	return
}

func srs_serve(c *net.TCPConn, writev bool, group [][]byte) (err error) {
	fmt.Println("serve client", c)

	// use write, send one by one packet.
	if (!writev) {
		for _,b := range group {
			if _,err = c.Write(b); err != nil {
				return
			}
		}
		return
	}

	// use writev to send group.
	if _,err = c.Writev(group); err != nil {
		return
	}
	return
}
