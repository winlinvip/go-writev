package main

import (
	"fmt"
	"os"
	"strconv"
	"net"
	"runtime"
	"runtime/pprof"
	"os/signal"
	"reflect"
	"syscall"
	"unsafe"
)

const(
	NbVideosInGroup = 512
	VideoSize = 4096
	HeaderSize = 12
)

func main() {
	var err error
	fmt.Println("golang version streaming server.")

	// always start profile.
	if w,err := os.Create("cpu.prof"); err != nil {
		panic(err)
	} else if err := pprof.StartCPUProfile(w); err != nil {
		panic(err)
	}

	// use signal kill to profile and quit.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func(){
		<- c

		pprof.StopCPUProfile()

		if w,err := os.Create("mem.prof"); err != nil {
			panic(err)
		} else if err := pprof.Lookup("heap").WriteTo(w, 0); err != nil {
			panic(err)
		}

		os.Exit(0)
	}()

	var port int
	var useWritev, writeOneByOne bool
	if len(os.Args) >= 3 {
		if port,err = strconv.Atoi(os.Args[1]); err != nil {
			panic(err)
		}
		useWritev = os.Args[2] == "true"
		writeOneByOne = false
	}
	if !useWritev && len(os.Args) >= 4 {
		writeOneByOne = os.Args[3] == "true"
	}
	if len(os.Args) < 3 || (!useWritev && len(os.Args) < 4) {
		fmt.Println("Usage:", os.Args[0], "<port> <use_writev> [write_one_by_one]")
		fmt.Println("   port: the tcp listen port.")
		fmt.Println("   use_writev: whether use writev. true or false.")
		fmt.Println("   write_one_by_one: for write(not writev), whether send packet one by one.")
		fmt.Println("Fox example:")
		fmt.Println("   ", os.Args[0], "1985 true")
		fmt.Println("   ", os.Args[0], "1985 false true")
		fmt.Println("   ", os.Args[0], "1985 false false")
		os.Exit(-1)
	}

	runtime.GOMAXPROCS(4)
	fmt.Println("always use 4 cpus")

	fmt.Println(fmt.Sprintf("listen at tcp://%v, use writev %v", port, useWritev))
	if (!useWritev) {
		fmt.Println("for write, send one-by-one", writeOneByOne)
	}

	var addr *net.TCPAddr
	if addr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%v", port)); err != nil {
		panic(err)
	}

	var listener *net.TCPListener
	if listener,err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		var conn *net.TCPConn
		if conn,err = listener.AcceptTCP(); err != nil {
			panic(err)
		}

		go func(c *net.TCPConn) {
			defer c.Close()

			c.SetNoDelay(false)

			// assume there is a video stream, which contains infinite video packets,
			// server must delivery all video packets to client.
			// for high performance, we send a group of video(to avoid too many syscall),
			// here we send 10 videos as a group.
			for {
				// @remark for test, each video is M bytes.
				video := make([]byte, VideoSize)

				// @remark for test, each video header is M0 bytes.
				header := make([]byte, HeaderSize)

				// @remark for test, each group contains N (header+video)s.
				group := make([][]byte, 2 * NbVideosInGroup)
				for i := 0; i < 2 * NbVideosInGroup; i += 2 {
					group[i] = header
					group[i + 1] = video
				}

				// sendout the video group.
				if err = srs_send(conn, group, useWritev, writeOneByOne); err != nil {
					panic(err)
				}
			}
		}(conn)
	}
}

// use write, to avoid lots of syscall, we copy to a big buffer.
var bigBuffer []byte = make([]byte, NbVideosInGroup * (HeaderSize + VideoSize))

// each group contains N (header+video)s.
//      header is M bytes.
//      videos is M0 bytes.
func srs_send(conn *net.TCPConn, group [][]byte, useWritev, writeOneByOne bool) (err error) {
	if (useWritev) {
		return writev(conn, group)
	}

	// use write, send one by one packet.
	// @remark avoid memory copy, but with lots of syscall, hurts performance.
	if (writeOneByOne) {
		for i := 0; i < 2 * NbVideosInGroup; i++ {
			if _,err = conn.Write(group[i]); err != nil {
				return
			}
		}
		return
	}

	var nn int
	for i := 0; i < 2 * NbVideosInGroup; i++ {
		b := group[i]
		copy(bigBuffer[nn:nn + len(b)], b)
		nn += len(b)
	}

	if _,err = conn.Write(bigBuffer); err != nil {
		return
	}
	return;
}

func writev(c *net.TCPConn, g [][]byte) (err error) {
	var fc reflect.Value
	if fc = reflect.ValueOf(c); fc.Kind() == reflect.Ptr {
		fc = fc.Elem()
	}

	var ffd reflect.Value = fc.FieldByName("fd")
	if ffd.Kind() == reflect.Ptr {
		ffd = ffd.Elem()
	}

	var fsysfd reflect.Value = ffd.FieldByName("sysfd")
	if fsysfd.Kind() == reflect.Ptr {
		fsysfd = fsysfd.Elem()
	}

	fd := uintptr(fsysfd.Int())
	//fmt.Println("fd is", fd)

	iovs := make([]syscall.Iovec, len(g))
	for i, iov := range g {
		iovs[i] = syscall.Iovec{&iov[0], uint64(len(iov))}
	}

	iovsPtr := uintptr(unsafe.Pointer(&iovs[0]))
	iovsLen := uintptr(len(iovs))

	_, _, e0 := syscall.Syscall(syscall.SYS_WRITEV, fd, iovsPtr, iovsLen)
	if e0 != 0 {
		if e0 != syscall.EAGAIN {
			panic(fmt.Sprintf("writev failed %v", e0))
		}
	}

	return
}
