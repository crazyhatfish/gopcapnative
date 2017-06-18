package main

import (
	"fmt"

	"github.com/crazyhatfish/gopcapnative"
)

func main() {
	fmt.Println("gopcapnative")

	device := "\\\\.\\Global\\NPF_{7B6F50F2-CA58-4CDD-9DB3-0DE6148E482E}"
	pcap, err := OpenLivePcap(device)
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
