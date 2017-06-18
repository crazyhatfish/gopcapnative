// +build linux

package gopcapnative

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

const (
	ETH_P_ALL_HTONS = 0x0300
)

func OpenLivePcap(device string) (*LivePcap, error) {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, ETH_P_ALL_HTONS)
	if err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByName(device)
	if err != nil {
		return nil, err
	}

	sockAddr := syscall.SockaddrLinklayer{}
	sockAddr.Ifindex = iface.Index
	sockAddr.Protocol = ETH_P_ALL_HTONS
	if err := syscall.Bind(fd, &sockAddr); err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))

	result := &LivePcap{
		device: device,
		handle: fd,
		file:   f,
	}

	if err := result.SetBufferSize(65536); err != nil {
		return nil, err
	}

	return result, nil
}

type LivePcap struct {
	device     string
	handle     int
	file       *os.File
	bufferSize uint32
}

func (this *LivePcap) SetBufferSize(size uint32) error {
	this.bufferSize = size
	return syscall.SetsockoptInt(this.handle, syscall.SOL_SOCKET, syscall.SO_RCVBUF, int(size))
}

func (this *LivePcap) Read() ([][]byte, error) {
	buffer := make([]byte, this.bufferSize)
	n, err := this.file.Read(buffer)
	if err != nil {
		return nil, err
	}

	// CHECK: multiple packets possible?
	return [][]byte{buffer[:n]}, nil
}

func (this *LivePcap) Close() {
	if err := this.file.Close(); err != nil {
		panic(err)
	}
	this.handle = -1
}
