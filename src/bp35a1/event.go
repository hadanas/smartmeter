package bp35a1

import (
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

type ev int

const (
	ESREG ev = iota
	EINFO
	EVER
	EAPPVER

	ERXUDP
	ERXTCP
	EPONG
	ETCP
	EADDR
	ENEIGHBOR
	EPANDESC
	EEDSCAN
	EPORT
	EHANDLE
	EVENT
)

type MultiLine interface {
	Parse([]string)
}

/* Event */
type Event interface {
	Type() ev
}

type event struct {
	t ev
}

func (e *event) Type() ev {
	return e.t
}

/* EventSreg */
type EventSreg interface {
	Event
	Val() string
}

type event_sreg struct {
	*event
	val string
}

func (e *event_sreg) Val() string {
	return e.val
}

/* EventInfo */
type EventInfo interface {
	Event
	IpAddr() net.IP
	HwAddr() string
	Channel() uint8
	PanId() uint16
	Addr16() uint16
}

type event_info struct {
	*event
	ipaddr  net.IP
	hwaddr  string
	channel uint8
	panid   uint16
	addr16  uint16
}

func (e *event_info) IpAddr() net.IP {
	return e.ipaddr
}

func (e *event_info) HwAddr() string {
	return e.hwaddr
}

func (e *event_info) Channel() uint8 {
	return e.channel
}

func (e *event_info) PanId() uint16 {
	return e.panid
}

func (e *event_info) Addr16() uint16 {
	return e.addr16
}

/* EventVer */
type EventVer interface {
	Event
	Version() string
}

type event_ver struct {
	*event
	version string
}

func (e *event_ver) Version() string {
	return e.version
}

/* EventAppVer */
type EventAppVer interface {
	Event
	Version() string
}

type event_appver struct {
	*event
	version string
}

func (e *event_appver) Version() string {
	return e.version
}

/* EventRxUDP */
type EventRxUDP interface {
	Event
	Sender() net.IP
	Dest() net.IP
	RPort() uint16
	LPort() uint16
	SenderLLA() string
	Secured() uint8
	Data() []byte
}

type event_rxudp struct {
	*event
	sender    net.IP
	dest      net.IP
	rport     uint16
	lport     uint16
	senderlla string
	secured   uint8
	data      []byte
}

func (e *event_rxudp) Sender() net.IP {
	return e.sender
}

func (e *event_rxudp) Dest() net.IP {
	return e.dest
}

func (e *event_rxudp) RPort() uint16 {
	return e.rport
}

func (e *event_rxudp) LPort() uint16 {
	return e.lport
}

func (e *event_rxudp) SenderLLA() string {
	return e.senderlla
}

func (e *event_rxudp) Secured() uint8 {
	return e.secured
}

func (e *event_rxudp) Data() []byte {
	return e.data
}

/* EventRxTCP */
type EventRxTCP interface {
	Event
	Sender() net.IP
	RPort() uint16
	LPort() uint16
	SenderLLA() string
	Data() []byte
}

type event_rxtcp struct {
	*event
	sender    net.IP
	rport     uint16
	lport     uint16
	senderlla string
	data      []byte
}

func (e *event_rxtcp) Sender() net.IP {
	return e.sender
}

func (e *event_rxtcp) RPort() uint16 {
	return e.rport
}

func (e *event_rxtcp) LPort() uint16 {
	return e.lport
}

func (e *event_rxtcp) SenderLLA() string {
	return e.senderlla
}

func (e *event_rxtcp) Data() []byte {
	return e.data
}

/* EventPong */
type EventPong interface {
	Event
	Sender() net.IP
}

type event_pong struct {
	*event
	sender net.IP
}

func (e *event_pong) Sender() net.IP {
	return e.sender
}

/* EventTCP */
type EventTCP interface {
	Event
	Status() uint8
	Handle() uint8
	IpAddr() net.IP
	RPort() uint16
	LPort() uint16
}

type event_tcp struct {
	*event
	status uint8
	handle uint8
	ipaddr net.IP
	rport  uint16
	lport  uint16
}

func (e *event_tcp) Status() uint8 {
	return e.status
}

func (e *event_tcp) Handle() uint8 {
	return e.handle
}

func (e *event_tcp) IpAddr() net.IP {
	return e.ipaddr
}

func (e *event_tcp) RPort() uint16 {
	return e.rport
}

func (e *event_tcp) LPort() uint16 {
	return e.lport
}

/* EventAddr */
type EventAddr interface {
	Event
	MultiLine
	IpAddrs() []net.IP
}

type event_addr struct {
	*event
	ipaddrs []net.IP
}

func (e *event_addr) IpAddrs() []net.IP {
	return e.ipaddrs
}

func (e *event_addr) Parse(data []string) {
	for _, d := range data {
		e.ipaddrs = append(e.ipaddrs, net.ParseIP(d))
	}
}

/* EventNeighbor */
type Neighbor interface {
	IpAddr() net.IP
	HwAddr() string
	Addr16() uint16
}

type neighbor struct {
	ipaddr net.IP
	hwaddr string
	addr16 uint16
}

func (n neighbor) IpAddr() net.IP {
	return n.ipaddr
}

func (n neighbor) HwAddr() string {
	return n.hwaddr
}

func (n neighbor) Addr16() uint16 {
	return n.addr16
}

type EventNeighbor interface {
	Event
	MultiLine
	Neighbors() []Neighbor
}

type event_neighbor struct {
	*event
	neighbors []*neighbor
}

func (e *event_neighbor) Neighbors() []Neighbor {
	n := make([]Neighbor, len(e.neighbors))
	for i, v := range e.neighbors {
		n[i] = v
	}
	return n
}

func (e *event_neighbor) Parse(data []string) {
	for _, d := range data {
		c := strings.Split(d, " ")
		e.neighbors = append(e.neighbors,
			&neighbor{
				ipaddr: net.ParseIP(c[0]),
				hwaddr: c[1],
				addr16: uint16(atoi(c[2]))})
	}
}

/* EventPanDesc */
type EventPanDesc interface {
	Event
	MultiLine
	Channel() uint8
	Page() uint8
	PanId() uint16
	Addr() string
	LQI() uint8
	PairId() string
}

type event_pandesc struct {
	*event
	channel uint8
	page    uint8
	panid   uint16
	addr    string
	lqi     uint8
	pairid  string
}

func (e *event_pandesc) Channel() uint8 {
	return e.channel
}

func (e *event_pandesc) Page() uint8 {
	return e.page
}

func (e *event_pandesc) PanId() uint16 {
	return e.panid
}

func (e *event_pandesc) Addr() string {
	return e.addr
}

func (e *event_pandesc) LQI() uint8 {
	return e.lqi
}

func (e *event_pandesc) PairId() string {
	return e.pairid
}

func (e *event_pandesc) Parse(data []string) {
	for _, d := range data {
		c := strings.Split(d, ":")
		name, value := c[0], c[1]

		switch name {
		case "  Channel":
			e.channel = uint8(atoi(value))
		case "  Channel Page":
			e.page = uint8(atoi(value))
		case "  Pan ID":
			e.panid = uint16(atoi(value))
		case "  Addr":
			e.addr = value
		case "  LQI":
			e.lqi = uint8(atoi(value))
		case "  PairID":
			e.pairid = value
		}
	}
}

/* EventEdScan */
type EdVal interface {
	Channel() uint8
	Rssi() uint8
}

type edval struct {
	channel uint8
	rssi    uint8
}

func (e edval) Channel() uint8 {
	return e.channel
}

func (e edval) Rssi() uint8 {
	return e.rssi
}

type EventEdScan interface {
	Event
	MultiLine
	EdVals() []EdVal
}

type event_edscan struct {
	*event
	edvals []edval
}

func (e *event_edscan) EdVals() []EdVal {
	d := make([]EdVal, len(e.edvals))
	for i, v := range e.edvals {
		d[i] = v
	}
	return d
}

func (e *event_edscan) Parse(data []string) {
	for _, d := range data {
		var c, r string
		s := strings.Split(d, " ")

		for len(s) > 0 {
			c, r, s = s[0], s[1], s[2:]
			e.edvals = append(e.edvals,
				edval{
					channel: uint8(atoi(c)),
					rssi:    uint8(atoi(r))})
		}
	}
}

/* EventPort */
type EventPort interface {
	Event
	MultiLine
	UdpPorts() [6]uint16
	TcpPorts() [4]uint16
}

type event_port struct {
	*event
	udpports [6]uint16
	tcpports [4]uint16
}

func (e *event_port) UdpPorts() [6]uint16 {
	return e.udpports
}

func (e *event_port) TcpPorts() [4]uint16 {
	return e.tcpports
}

func (e *event_port) Parse(data []string) {
	for i := 0; i < 6; i = i + 1 {
		e.udpports[i] = uint16(atoi(data[i]))
	}
	for i := 0; i < 4; i = i + 1 {
		e.tcpports[i] = uint16(atoi(data[i+7]))
	}
}

/* EventHandle */
type Handle interface {
	Handle() uint8
	IpAddr() net.IP
	RPort() uint16
	LPort() uint16
}

type handle struct {
	handle uint8
	ipaddr net.IP
	rport  uint16
	lport  uint16
}

func (h handle) Handle() uint8 {
	return h.handle
}

func (h handle) IpAddr() net.IP {
	return h.ipaddr
}

func (h handle) RPort() uint16 {
	return h.rport
}

func (h handle) LPort() uint16 {
	return h.lport
}

type EventHandle interface {
	Event
	MultiLine
	Handles() []Handle
}

type event_handle struct {
	*event
	handles []handle
}

func (e *event_handle) Handles() []Handle {
	h := make([]Handle, len(e.handles))
	for i, v := range e.handles {
		h[i] = v
	}
	return h
}

func (e *event_handle) Parse(data []string) {
	for _, d := range data {
		c := strings.Split(d, " ")
		e.handles = append(e.handles,
			handle{
				handle: uint8(atoi(c[0])),
				ipaddr: net.ParseIP(c[1]),
				rport:  uint16(atoi(c[2])),
				lport:  uint16(atoi(c[3]))})
	}
}

/* EventEvent */
type EventEvent interface {
	Event
	Num() uint8
	Sender() net.IP
	Param() []byte
}

type event_event struct {
	*event
	num    uint8
	sender net.IP
	param  []byte
}

func (e *event_event) Num() uint8 {
	return e.num
}

func (e *event_event) Sender() net.IP {
	return e.sender
}

func (e *event_event) Param() []byte {
	return e.param
}

func atoi(s string) int {
	i64, _ := strconv.ParseInt(s, 16, 0)
	return int(i64)
}

func newEvent(data string) Event {
	d := strings.Split(data, " ")

	e := &event{t: toEv(d[0])}
	switch e.t {
	case ESREG:
		return &event_sreg{
			event: e,
			val:   d[1]}
	case EINFO:
		return &event_info{
			event:   e,
			ipaddr:  net.ParseIP(d[1]),
			hwaddr:  d[2],
			channel: uint8(atoi(d[3])),
			panid:   uint16(atoi(d[4])),
			addr16:  uint16(atoi(d[5]))}
	case EVER:
		return &event_ver{
			event:   e,
			version: d[1]}
	case EAPPVER:
		return &event_appver{
			event:   e,
			version: d[1]}
	case ERXUDP:
		return &event_rxudp{
			event:     e,
			sender:    net.ParseIP(d[1]),
			dest:      net.ParseIP(d[2]),
			rport:     uint16(atoi(d[3])),
			lport:     uint16(atoi(d[4])),
			senderlla: d[5],
			secured:   uint8(atoi(d[6])),
			data: func() []byte {
				h, _ := hex.DecodeString(d[8])
				return h
			}()}
	case ERXTCP:
		return &event_rxtcp{
			event:     e,
			sender:    net.ParseIP(d[1]),
			rport:     uint16(atoi(d[2])),
			lport:     uint16(atoi(d[3])),
			senderlla: d[4],
			data: func() []byte {
				h, _ := hex.DecodeString(d[6])
				return h
			}()}
	case EPONG:
		return &event_pong{
			event:  e,
			sender: net.ParseIP(d[1])}
	case ETCP:
		s := uint8(atoi(d[1]))
		if s == 1 {
			return &event_tcp{
				event:  e,
				status: s,
				handle: uint8(atoi(d[2]))}
		} else {
			return &event_tcp{
				event:  e,
				status: s,
				handle: uint8(atoi(d[2])),
				ipaddr: net.ParseIP(d[3]),
				rport:  uint16(atoi(d[4])),
				lport:  uint16(atoi(d[5]))}
		}
	case EADDR:
		return &event_addr{
			event:   e,
			ipaddrs: []net.IP{}}
	case ENEIGHBOR:
		return &event_neighbor{
			event:     e,
			neighbors: []*neighbor{}}
	case EPANDESC:
		return &event_pandesc{
			event: e}
	case EEDSCAN:
		return &event_edscan{
			event:  e,
			edvals: []edval{}}
	case EPORT:
		return &event_port{
			event:    e,
			udpports: [6]uint16{},
			tcpports: [4]uint16{}}
	case EHANDLE:
		return &event_handle{
			event:   e,
			handles: []handle{}}
	case EVENT:
		n := uint8(atoi(d[1]))
		if n == 0x21 {
			return &event_event{
				event:  e,
				num:    n,
				sender: net.ParseIP(d[2]),
				param: func() []byte {
					h, _ := hex.DecodeString(d[3])
					return h
				}()}
		} else {
			return &event_event{
				event:  e,
				num:    n,
				sender: net.ParseIP(d[2])}
		}
	}

	return e
}
