package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/crazyhatfish/gopcapnative"
)

func main() {
	fmt.Println("gopcapnative")

	if len(os.Args) < 2 ||
		(len(os.Args[1]) > 0 && os.Args[1][0] == '-') {
		fmt.Printf("usage: %s <device>\n", os.Args[0])
		return
	}

	// "\\\\.\\Global\\NPF_{7B6F50F2-CA58-4CDD-9DB3-0DE6148E482E}"
	// "wlx801f02593400"
	pcap, err := gopcapnative.OpenLivePcap(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer pcap.Close()
	pcap.SetBufferSize(65536)

	for {
		bufs, err := pcap.Read()
		if err != nil {
			panic(err)
		}
		for _, buf := range bufs {
			if len(buf) > 0 {
				fmt.Println(hex.EncodeToString(buf))
			}
		}
	}
}
