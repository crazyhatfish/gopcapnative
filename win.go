// +build windows

package gopcapnative

import (
	"golang.org/x/sys/windows/registry"

	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	// XXX: do we need this??
	var d syscall.WSAData
	if err := syscall.WSAStartup(uint32(0x202), &d); err != nil {
		panic(err)
	}
}

type Adapter struct {
	Name string
	Path string
}

func getAdapters() ([]Adapter, error) {
	result := []Adapter{}

	basePath := "SYSTEM\\CurrentControlSet\\Control\\Class\\{4D36E972-E325-11CE-BFC1-08002BE10318}"

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, basePath, registry.ENUMERATE_SUB_KEYS|registry.READ|registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(1024)
	if err != nil && err != io.EOF {
		return nil, err
	}
	for _, name := range names {
		if _, err := strconv.Atoi(name); err != nil {
			continue
		}

		adapterK, err := registry.OpenKey(k, name, registry.READ)
		if err != nil {
			return nil, err
		}
		subNames, err := adapterK.ReadSubKeyNames(10)
		if err != nil && err != io.EOF {
			return nil, err
		}
		found := false
		for _, subName := range subNames {
			if strings.ToUpper(subName) == "NDI" {
				found = true
			}
		}
		if !found {
			continue
		}

		desc, _, err := adapterK.GetStringValue("DriverDesc")
		if err != nil {
			return nil, err
		}
		// HACK: HACK!
		if strings.Contains(desc, "ISA") || strings.Contains(desc, "6to4") ||
			strings.Contains(desc, "Miniport") || strings.Contains(desc, "RAS Async") {
			continue
		}
		uuid, _, err := adapterK.GetStringValue("NetCfgInstanceId")
		if err != nil {
			return nil, err
		}
		path := "\\\\.\\Global\\NPF_" + uuid

		result = append(result, Adapter{desc, path})
	}

	return result, nil
}

const (
	BIOCSETBUFFERSIZE = 9592
	BIOCSMINTOCOPY    = 7414
	BIOCSETOID        = 2147483648

	OID_GEN_CURRENT_PACKET_FILTER = 0x0001010E

	NDIS_PACKET_TYPE_PROMISCUOUS = 0x0020
)

func OpenLivePcap(device string) (*LivePcap, error) {
	devicep, err := syscall.UTF16PtrFromString(device)
	if err != nil {
		return nil, err
	}

	handle, err := syscall.CreateFile(devicep, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, 0, 0)
	if err != nil {
		return nil, err
	}

	result := &LivePcap{
		device: device,
		handle: handle,
	}

	result.SetHwFilter(NDIS_PACKET_TYPE_PROMISCUOUS)
	result.setIntParam(BIOCSMINTOCOPY, 0)
	result.SetBufferSize(65536)

	return result, nil
}

type LivePcap struct {
	device     string
	handle     syscall.Handle
	bufferSize uint32
}

func (this *LivePcap) setIntParam(ioControlCode uint32, val uint32) {
	var buffer [4]byte
	var bytesReturned uint32
	binary.LittleEndian.PutUint32(buffer[:], val)
	syscall.DeviceIoControl(this.handle, ioControlCode, &buffer[0], uint32(len(buffer)), nil, 0, &bytesReturned, nil)
}

func (this *LivePcap) SetBufferSize(size uint32) error {
	this.bufferSize = size
	this.setIntParam(BIOCSETBUFFERSIZE, size)
	return nil
}

func (this *LivePcap) SetHwFilter(filter uint32) {
	var buffer [15]byte
	var bytesReturned uint32
	binary.LittleEndian.PutUint32(buffer[:4], OID_GEN_CURRENT_PACKET_FILTER)
	binary.LittleEndian.PutUint32(buffer[4:8], 4)
	binary.LittleEndian.PutUint32(buffer[8:12], filter)
	syscall.DeviceIoControl(this.handle, BIOCSETOID, &buffer[0], uint32(len(buffer)), &buffer[0], uint32(len(buffer)), &bytesReturned, nil)
}

func (this *LivePcap) Read() ([][]byte, error) {
	type BpfHdr struct {
		Timestamp uint32
		Unknown   uint32 // XXX: seriously what is this??
		CapLen    uint32
		DataLen   uint32
		HdrLen    uint16
		Padding   uint16 // XXX: sometimes not empty, why??
	}

	done := uint32(this.bufferSize)
	buffer := make([]byte, this.bufferSize)
	err := syscall.ReadFile(this.handle, buffer, &done, nil)
	if err != nil {
		return nil, err
	}

	result := [][]byte{}
	header := BpfHdr{}
	reader := bytes.NewBuffer(buffer[:done])
	for reader.Len() > 0 {
		binary.Read(reader, binary.LittleEndian, &header)
		pktSize := header.DataLen + uint32(header.HdrLen)
		paddingSize := 4 - (pktSize % 4)
		if paddingSize == 4 {
			paddingSize = 0
		}
		pkt := reader.Next(int(header.DataLen))
		reader.Next(int(paddingSize))
		if header.DataLen > 0 {
			result = append(result, pkt)
		}
	}

	return result, nil
}

func (this *LivePcap) Close() {
	syscall.CloseHandle(this.handle)
}
