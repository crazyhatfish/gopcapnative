# gopcapnative
portable pure Go packet sniffing with pcap-like API

## Dependencies

### Windows:
- [WinPCAP](https://www.winpcap.org/install/default.htm) 4.1.3 (may work with earlier versions)
- `go get` [`golang.org/x/sys/windows/registry`](golang.org/x/sys/windows/registry)

## Notes

### Linux:
- Only supports `PF_PACKET`

## todo
- [x] Support Windows (with WinPCAP)
- [x] Turn into a library
- [ ] Support npcap
- [x] Support Linux
- [ ] Support Mac
- [ ] Support wireless monitor mode
