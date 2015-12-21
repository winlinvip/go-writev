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
	"time"
	"runtime/debug"
	"bufio"
)

var bytesWritten uint64

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
	var tcpNoDelay = flag.Bool("nodelay", false, "set TCP_NODELAY for connection.")
	var traceInterval = flag.Int("trace", 3, "the trace interval in seconds.")

	flag.Usage = func(){
		fmt.Println(fmt.Sprintf("Usage: %v [--port=int] [--writev=bool] [--cpus=int] [--cpup=string] [--memp=string] [--group=int] [--header=int] [--payload=int] [--trace=int] [-h|--help]", os.Args[0]))
		fmt.Println(fmt.Sprintf("	port, the listen port. default 1985"))
		fmt.Println(fmt.Sprintf("	writev, whether use writev to send. default false(use write)"))
		fmt.Println(fmt.Sprintf("	cpus, the cpus to use. default 1"))
		fmt.Println(fmt.Sprintf("	cpup, the cpu profile file. default cpu.prof"))
		fmt.Println(fmt.Sprintf("	memp, the memory profile file. default mem.prof"))
		fmt.Println(fmt.Sprintf("	group, the size of group, which contains N*(header, payload). default 512"))
		fmt.Println(fmt.Sprintf("	header, the size of header. default 12"))
		fmt.Println(fmt.Sprintf("	payload, the size of payload. default 1024"))
		fmt.Println(fmt.Sprintf("	nodelay, set the TCP_NODELAY to. default false"))
		fmt.Println(fmt.Sprintf("	trace, the trace interval in seconds. default 3"))
		fmt.Println(fmt.Sprintf("	help, show this help and exit"))
		fmt.Println(fmt.Sprintf("Remark:"))
		fmt.Println(fmt.Sprintf("	use SIGINT or ctrl+c to interrupt program and collect cpu/mem profile data."))
		fmt.Println(fmt.Sprintf("For example:"))
		fmt.Println(fmt.Sprintf("	%v --port=1985 --writev=false", os.Args[0]))
	}
	flag.Parse()

	runtime.GOMAXPROCS(*nbCpus)
	fmt.Println(fmt.Sprintf("tcp server, writev:%v, listen:%v, cpus:%v, nodelay:%v, cpuProfile:%v, memProfile:%v, group:%v, header:%v, payload:%v",
		*useWritev, *port, *nbCpus, *tcpNoDelay, *cpuProfile, *memProfile, *groupSize, *headerSize, *payloadSize))
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

	// trace the performance.
	go func() {
		var previous uint64
		sample := time.Now()
		stat := &debug.GCStats{}
		for {
			diff := bytesWritten - previous
			elapse := time.Now().Sub(sample) / time.Millisecond
			if (diff > 0 && elapse > 0) {
				mbps := float64(diff) * 8 / float64(elapse) / 1000
				debug.ReadGCStats(stat)
				if len(stat.Pause) > 3 {
					stat.Pause = append([]time.Duration{}, stat.Pause[:3]...)
				}
				fmt.Println(fmt.Sprintf("%.2fMbps, gc %v %v", mbps, stat.NumGC, stat.Pause))
			}
			previous = bytesWritten
			sample = time.Now()

			time.Sleep(time.Second * time.Duration(*traceInterval))
		}
	} ()

	var err error
	var l *net.TCPListener
	var addr *net.TCPAddr
	if addr,err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%v", *port)); err != nil {
		panic(err)
	}
	if l,err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}
	fmt.Println("listen at", fmt.Sprintf("tcp://%v", addr))

	// collect the bytes written.
	cnn := make(chan int, 64)
	go func(){
		for nn := range cnn {
			bytesWritten += uint64(nn)
		}
	}()

	// let all GC out of this benchmark.
	group := srs_recv_group_packets(groupSize, headerSize, payloadSize)

	for {
		var c *net.TCPConn
		if c,err = l.AcceptTCP(); err != nil {
			panic(err)
		}
		// support concurrency for each client.
		go func() {
			if err = srs_serve(c, *tcpNoDelay, *useWritev, group, cnn); err != nil {
				fmt.Println("ignore serve client err", err)
			}
		}()
	}
}

func srs_serve(c *net.TCPConn, nodelay, writev bool, group [][]byte, cnn chan int) (err error) {
	c.SetNoDelay(nodelay)
	fmt.Println(fmt.Sprintf("serve %v with TCP_NODELAY:%v", c.RemoteAddr(), nodelay))
	defer fmt.Println(fmt.Sprintf("server %v ok", c.RemoteAddr()))

	var nn int
	for _,b := range group {
		nn += len(b)
	}

	// use write, send one by one packet.
	if (!writev) {
		// use bufio to cache and flush the group.
		w := bufio.NewWriterSize(c, nn)
		// write to bufio and flush iovecs.
		for {
			for _,b := range group {
				if _,err = w.Write(b); err != nil {
					return
				}
			}
			if err = w.Flush(); err != nil {
				return
			}
			cnn <- nn
		}
		return
	}

	// use writev to send group.
	for {
		if _,err = c.Writev(group); err != nil {
			return
		}
		cnn <- nn
	}
	return
}

func srs_recv_group_packets(groupSize, headerSize, payloadSize int) ([][]byte) {
	// create group of tcp packets to send.
	// each group is a serial of header and payload,
	// 		group = h0, p0, h1, p1, ..., hN, pN
	group := make([][]byte, groupSize * 2)
	for i := 0; i < len(group); i += 2 {
		// create header.
		header := make([]byte, headerSize)
		// create payload
		payload := make([]byte, payloadSize)
		// the group contains N*(header, payload)
		// h0, p0, h1, p1, ..., hN, pN
		group[i] = header
		group[i+1]=payload
	}

	return group
}
