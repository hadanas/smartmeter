package bp35a1

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

type cmd int

const (
	SKSREG cmd = iota
	SKINFO
	SKSTART
	SKJOIN
	SKREJOIN
	SKTERM
	SKSENDTO
	SKCONNECT
	SKSEND
	SKCLOSE
	SKPING
	SKSCAN
	SKREGDEV
	SKRMDEV
	SKSETKEY
	SKRMKEY
	SKSECENABLE
	SKSETPSK
	SKSETPWD
	SKSETRBID
	SKADDNBR
	SKUDPPORT
	SKTCPPORT
	SKSAVE
	SKLOAD
	SKERASE
	SKVER
	SKAPPVER
	SKRESET
	SKTABLE
	SKDSLEEP
	SKRFLO
	SKLL64
)

type Command interface {
	String() string
	Parameters() []interface{}
}

type command struct {
	id cmd
}

func (c *command) String() string {
	return c.id.String()
}

func (c *command) Parameters() []interface{} {
	return []interface{}{}
}

type command_sreg struct {
	*command
	reg uint8
	val string
}

func (c *command_sreg) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("S%02X", c.reg),
		c.val}
}

type command_join struct {
	*command
	ipaddr net.IP
}

func (c *command_join) Parameters() []interface{} {
	return []interface{}{iptoa(c.ipaddr)}
}

type command_sendto struct {
	*command
	handle uint8
	ipaddr net.IP
	port   uint16
	sec    uint8
	data   []byte
}

func (c *command_sendto) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%d", c.handle),
		iptoa(c.ipaddr),
		fmt.Sprintf("%04X", c.port),
		fmt.Sprintf("%d", c.sec),
		fmt.Sprintf("%04X", len(c.data)),
		c.data}
}

type command_connect struct {
	*command
	ipaddr net.IP
	rport  uint16
	lport  uint16
}

func (c *command_connect) Parameters() []interface{} {
	return []interface{}{
		c.ipaddr.String(),
		fmt.Sprintf("%04X", c.rport),
		fmt.Sprintf("%04X", c.lport)}
}

type command_send struct {
	*command
	handle uint8
	data   []byte
}

func (c *command_send) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", c.handle),
		fmt.Sprintf("%04X", len(c.data)),
		c.data}
}

type command_close struct {
	*command
	handle uint8
}

func (c *command_close) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", c.handle)}
}

type command_ping struct {
	*command
	ipaddr net.IP
}

func (c *command_ping) Parameters() []interface{} {
	return []interface{}{iptoa(c.ipaddr)}
}

type command_scan struct {
	*command
	mode     uint8
	mask     uint32
	duration uint8
}

func (c *command_scan) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%d", c.mode),
		fmt.Sprintf("%08X", c.mask),
		fmt.Sprintf("%d", c.duration)}
}

type command_regdev struct {
	*command
	ipaddr net.IP
}

func (c *command_regdev) Parameters() []interface{} {
	return []interface{}{iptoa(c.ipaddr)}
}

type command_rmdev struct {
	*command
	ipaddr net.IP
}

func (c *command_rmdev) Parameters() []interface{} {
	return []interface{}{iptoa(c.ipaddr)}
}

type command_setkey struct {
	*command
	index uint8
	key   []byte
}

func (c *command_setkey) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", c.index),
		strings.ToUpper(hex.EncodeToString(c.key))}
}

type command_rmkey struct {
	*command
	index uint8
}

func (c *command_rmkey) Parameters() []interface{} {
	return []interface{}{fmt.Sprintf("%02X", c.index)}
}

type command_secenable struct {
	*command
	mode   int
	ipaddr net.IP
	hwaddr string
}

func (c *command_secenable) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%04X", c.mode),
		iptoa(c.ipaddr),
		c.hwaddr}
}

type command_setpsk struct {
	*command
	key []byte
}

func (c *command_setpsk) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", len(c.key)),
		strings.ToUpper(hex.EncodeToString(c.key))}
}

type command_setpwd struct {
	*command
	pwd string
}

func (c *command_setpwd) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", len(c.pwd)),
		c.pwd}
}

type command_setrbid struct {
	*command
	rbid string
}

func (c *command_setrbid) Parameters() []interface{} {
	return []interface{}{c.rbid}
}

type command_addnbr struct {
	*command
	ipaddr net.IP
	hwaddr string
}

func (c *command_addnbr) Parameters() []interface{} {
	return []interface{}{
		iptoa(c.ipaddr),
		c.hwaddr}
}

type command_udpport struct {
	*command
	handle uint8
	port   uint16
}

func (c *command_udpport) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", c.handle),
		fmt.Sprintf("%04X", c.port)}
}

type command_tcpport struct {
	*command
	index uint8
	port  uint16
}

func (c *command_tcpport) Parameters() []interface{} {
	return []interface{}{
		fmt.Sprintf("%02X", c.index),
		fmt.Sprintf("%04X", c.port)}
}

type command_table struct {
	*command
	mode uint8
}

func (c *command_table) Parameters() []interface{} {
	return []interface{}{fmt.Sprintf("%02X", c.mode)}
}

type command_rflo struct {
	*command
	mode uint8
}

func (c *command_rflo) Parameters() []interface{} {
	return []interface{}{fmt.Sprintf("%02X", c.mode)}
}

type command_ll64 struct {
	*command
	hwaddr string
}

func (c *command_ll64) Parameters() []interface{} {
	return []interface{}{strings.ToUpper(c.hwaddr)}
}

func NewCommand(id cmd, params ...interface{}) Command {
	c := &command{id: id}
	switch id {
	case SKSREG:
		return &command_sreg{
			command: c,
			reg:     params[0].(uint8),
			val:     params[1].(string)}
	case SKJOIN:
		return &command_join{
			command: c,
			ipaddr:  params[0].(net.IP)}
	case SKSENDTO:
		return &command_sendto{
			command: c,
			handle:  params[0].(uint8),
			ipaddr:  params[1].(net.IP),
			port:    params[2].(uint16),
			sec:     params[3].(uint8),
			data:    params[4].([]byte)}
	case SKCONNECT:
		return &command_connect{
			command: c,
			ipaddr:  params[0].(net.IP),
			rport:   params[1].(uint16),
			lport:   params[2].(uint16)}
	case SKSEND:
		return &command_send{
			command: c,
			handle:  params[0].(uint8),
			data:    params[1].([]byte)}
	case SKCLOSE:
		return &command_close{
			command: c,
			handle:  params[0].(uint8)}
	case SKPING:
		return &command_ping{
			command: c,
			ipaddr:  params[0].(net.IP)}
	case SKSCAN:
		return &command_scan{
			command:  c,
			mode:     params[0].(uint8),
			mask:     params[1].(uint32),
			duration: params[2].(uint8)}
	case SKREGDEV:
		return &command_regdev{
			command: c,
			ipaddr:  params[0].(net.IP)}
	case SKRMDEV:
		return &command_rmdev{
			command: c,
			ipaddr:  params[0].(net.IP)}
	case SKSETKEY:
		return &command_setkey{
			command: c,
			index:   params[0].(uint8),
			key:     params[1].([]byte)}
	case SKRMKEY:
		return &command_rmkey{
			command: c,
			index:   params[0].(uint8)}
	case SKSECENABLE:
		return &command_secenable{
			command: c,
			mode:    params[0].(int),
			ipaddr:  params[1].(net.IP),
			hwaddr:  params[2].(string)}
	case SKSETPSK:
		return &command_setpsk{
			command: c,
			key:     params[0].([]byte)}
	case SKSETPWD:
		return &command_setpwd{
			command: c,
			pwd:     params[0].(string)}
	case SKSETRBID:
		return &command_setrbid{
			command: c,
			rbid:    params[0].(string)}
	case SKADDNBR:
		return &command_addnbr{
			command: c,
			ipaddr:  params[0].(net.IP),
			hwaddr:  params[1].(string)}
	case SKUDPPORT:
		return &command_udpport{
			command: c,
			handle:  params[0].(uint8),
			port:    params[1].(uint16)}
	case SKTCPPORT:
		return &command_tcpport{
			command: c,
			index:   params[0].(uint8),
			port:    params[1].(uint16)}
	case SKTABLE:
		return &command_table{
			command: c,
			mode:    params[0].(uint8)}
	case SKRFLO:
		return &command_rflo{
			command: c,
			mode:    params[0].(uint8)}
	case SKLL64:
		return &command_ll64{
			command: c,
			hwaddr:  params[1].(string)}
	}
	return c
}

func iptoa(ip net.IP) string {
	var v bytes.Buffer
	bt := []byte(ip)
	for i, b := range bt {
		if i > 0 && i%2 == 0 {
			v.WriteString(":")
		}
		v.WriteString(fmt.Sprintf("%02X", b))
	}
	return v.String()
}

func ToBytes(c Command) []byte {
	var buf []byte = []byte(c.String())

	for _, p := range c.Parameters() {
		buf = append(buf, 0x20)
		switch t := p.(type) {
		case []byte:
			buf = append(buf, t...)
		case string:
			buf = append(buf, []byte(t)...)
		default:
			s, ok := p.(fmt.Stringer)
			if ok {
				buf = append(buf, []byte(s.String())...)
			}
		}
	}
	return buf
}
