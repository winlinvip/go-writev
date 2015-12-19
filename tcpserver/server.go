// the prototype for go-writev,
// https://github.com/golang/go/issues/13451
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("the tcp server prototype for go-writev")
	fmt.Println("please read https://github.com/golang/go/issues/13451")

	if len(os.Args) < 2 {
		fmt.Println(fmt.Sprintf(`Usage: %v <write|writev>
	write, use multiple writes to send tcp packets.
	writev, use writev to send tcp packets.
For example:
	%v write
	%v writev`,
			os.Args[0], os.Args[0], os.Args[0]))
		os.Exit(-1)
	}

	var useWritev bool = (os.Args[1] == "writev")
	fmt.Println("to send group of tcp packets, writev:", useWritev)
}