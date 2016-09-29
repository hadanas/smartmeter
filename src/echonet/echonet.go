package echonet

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Esv byte

const (
	ESV_SETI        Esv = 0x60
	ESV_SETC        Esv = 0x61
	ESV_GET         Esv = 0x62
	ESV_INF_REQ     Esv = 0x63
	ESV_SET_RES     Esv = 0x71
	ESV_GET_RES     Esv = 0x72
	ESV_INF         Esv = 0x73
	ESV_INFC        Esv = 0x74
	ESV_INFC_RES    Esv = 0x7A
	ESV_SETI_SNA    Esv = 0x50
	ESV_SETC_SNA    Esv = 0x51
	ESV_GET_SNA     Esv = 0x52
	ESV_INF_SNA     Esv = 0x53
	ESV_SET_NO_RES  Esv = 0x70
	ESV_SET_GET     Esv = 0x6E
	ESV_SET_GET_RES Esv = 0x7E
	ESV_SET_GET_SNA Esv = 0x5E
)

type Class uint16

const (
	CLASS_SMART_EE_METER Class = 0x0288
	CLASS_CONTROLLER     Class = 0x05FF
)

type Epc byte

const (
	EPC_0288_EE_METER_CLASS                Epc = 0xD0
	EPC_0288_COMP_TRANS_RATIO              Epc = 0xD3
	EPC_0288_EFFECTIVE_DIGITS              Epc = 0xD7
	EPC_0288_CM_AMTS_OF_EE_NDIR            Epc = 0xE0
	EPC_0288_UNIT_FOR_CM_AMTS_OF_EE        Epc = 0xE1
	EPC_0288_HDATA_OF_CM_AMTS_OF_EE_NDIR   Epc = 0xE2
	EPC_0288_CM_AMTS_OF_EE_RDIR            Epc = 0xE3
	EPC_0288_HDATA_OF_CM_AMTS_OF_EE_RDIR   Epc = 0xE4
	EPC_0288_DAY_FOR_HDATA                 Epc = 0xE5
	EPC_0288_INST_EE                       Epc = 0xE7
	EPC_0288_INST_CURRENTS                 Epc = 0xE8
	EPC_0288_INST_VOLTAGES                 Epc = 0xE9
	EPC_0288_CM_AMTS_OF_EE_AT_FT_NDIR      Epc = 0xEA
	EPC_0288_CM_AMTS_OF_EE_AT_FT_RDIR      Epc = 0xEB
	EPC_0288_HDATA_OF_CM_AMTS_OF_EE_RDIR_2 Epc = 0xEC
	EPC_0288_DAY_FOR_HDATA_2               Epc = 0xED

	EPC_0EF0_OPERATING_STATUS           Epc = 0x80
	EPC_0EF0_VERSION_INFO               Epc = 0x82
	EPC_0EF0_ID_NUM                     Epc = 0x83
	EPC_0EF0_FAULT_CONTENT              Epc = 0x89
	EPC_0EF0_UNIQUE_ID_DATA             Epc = 0xBF
	EPC_0EF0_NUM_OF_SELF_NODE_INSTANCES Epc = 0xD3
	EPC_0EF0_NUM_OF_SELF_NODE_CLASSES   Epc = 0xD4
	EPC_0EF0_INSTANCE_LIST_NOTIFICATION Epc = 0xD5
	EPC_0EF0_SELF_NODE_INSTANCE_LIST_S  Epc = 0xD6
	EPC_0EF0_SELF_NODE_CLASS_LIST       Epc = 0xD7
)

type Encoder interface {
	Encode(uint16) []byte
}

type Decoder interface {
	Decode([]byte) Frame
}

type Frame interface {
	Encoder
	Decoder

	Seoj() (Class, uint8)
	Deoj() (Class, uint8)
	Esv() Esv
	Opc() uint8
	Properties() []Property

	SetSeoj(Class, uint8)
	SetDeoj(Class, uint8)
	SetEsv(Esv)
	SetOpc(uint8)
	SetProperties([]Property)
}

type Property interface {
	Epc() Epc
	Pdc() byte
	Edt() []byte

	SetEpc(Epc)
	SetPdc(byte)
	SetEdt([]byte)
}

type property struct {
	epc Epc
	pdc uint8
	edt []byte
}

func NewProperty() Property {
	return &property{}
}

func (p *property) Epc() Epc {
	return p.epc
}

func (p *property) Pdc() byte {
	return p.pdc
}

func (p *property) Edt() []byte {
	return p.edt
}

func (p *property) SetEpc(epc Epc) {
	p.epc = epc
}

func (p *property) SetPdc(n byte) {
	p.pdc = n
}

func (p *property) SetEdt(data []byte) {
	p.edt = data
}

type frame struct {
	seoj  Class
	seoji uint8
	deoj  Class
	deoji uint8
	esv   Esv
	opc   uint8
	props []*property
}

func NewFrame() Frame {
	return &frame{}
}

var header = []byte{0x10, 0x81}

func (f *frame) Encode(t uint16) []byte {
	buf := bytes.NewBuffer(header)
	write(buf, t)

	write(buf, f.seoj)
	write(buf, f.seoji)
	write(buf, f.deoj)
	write(buf, f.deoji)

	write(buf, f.esv)
	write(buf, f.opc)

	for i := 0; i < int(f.opc); i++ {
		write(buf, f.props[i].epc)
		write(buf, f.props[i].pdc)
		if f.props[i].pdc > 0 {
			write(buf, f.props[i].edt)
		}
	}

	return buf.Bytes()
}

func (f *frame) Decode(b []byte) Frame {
	buf := bytes.NewBuffer(b)

	buf.Next(4) // Header

	read(buf, &f.seoj)
	read(buf, &f.seoji)
	read(buf, &f.deoj)
	read(buf, &f.deoji)

	read(buf, &f.esv)
	read(buf, &f.opc)

	f.props = make([]*property, f.opc)
	for i := 0; i < int(f.opc); i++ {
		f.props[i] = &property{}
		read(buf, &f.props[i].epc)
		read(buf, &f.props[i].pdc)
		f.props[i].edt = buf.Next(int(f.props[i].pdc))
	}
	return f
}

func (f *frame) Seoj() (Class, uint8) {
	return f.seoj, f.seoji
}

func (f *frame) Deoj() (Class, uint8) {
	return f.deoj, f.deoji
}

func (f *frame) Esv() Esv {
	return f.esv
}

func (f *frame) Opc() uint8 {
	return f.opc
}

func (f *frame) Properties() []Property {
	p := make([]Property, len(f.props))
	for i, v := range f.props {
		p[i] = v
	}
	return p
}

func (f *frame) SetSeoj(class Class, index uint8) {
	f.seoj = class
	f.seoji = index
}

func (f *frame) SetDeoj(class Class, index uint8) {
	f.deoj = class
	f.deoji = index
}

func (f *frame) SetEsv(esv Esv) {
	f.esv = esv
}

func (f *frame) SetOpc(n uint8) {
	f.opc = n
}

func (f *frame) SetProperties(props []Property) {
	p := make([]*property, len(props))
	for i, v := range props {
		p[i] = v.(*property)
	}
	f.props = p
}

func read(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.BigEndian, data)
}

func write(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}
