package main

import (
	"flag"
	"fmt"

	lsrv "github.com/jaym/lsrv"
)

var (
	VERSION string = "0.0.1"
	socket         = flag.String("socket", "/tmp/lsrv.sock", "UNIX domain socket")
	daemon         = flag.Bool("daemon", false, "Run the lsrv daemon if true")
)

func main() {
	flag.Parse()

	if *daemon {
		fmt.Println("Starting lsrv daemon")
		lsrv.Spawn(*socket)
	} else {
		fmt.Println("Client mode")
	}
}
