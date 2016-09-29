package main

import (
	bp "bp35a1"
	"bytes"
	"echonet"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/jochenvg/go-udev"
	"github.com/robfig/cron"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/cihub/seelog"
)

var tranId = uint16(0)
var m = new(sync.Mutex)

func getTranId() uint16 {
	m.Lock()
	defer m.Unlock()

	tranId++
	return tranId
}

func main() {
	var path = flag.String("c", "smartmeter.conf", "config file")
	flag.Parse()

	var conf config
	loadConfig(*path, &conf)

	configLogger(conf.Log.Level)

	tty, err := getTTYPath()
	if err != nil {
		log.Critical(err)
		return
	}

	ctrl := bp.NewController(tty)

	ctrl.Send(bp.NewCommand(bp.SKSETPWD, conf.RouteB.Pwd))
	ctrl.Send(bp.NewCommand(bp.SKSETRBID, conf.RouteB.Id))

	var pan bp.EventPanDesc
	for pan == nil {
		ctrl.Send(bp.NewCommand(bp.SKSCAN, uint8(2), uint32(0xffffffff), uint8(6)),
			func(e bp.Event) bool {
				switch e.Type() {
				case bp.EPANDESC:
					pan = e.(bp.EventPanDesc)
				case bp.EVENT:
					return e.(bp.EventEvent).Num() == 0x22
				}
				return false
			})
	}

	ctrl.Send(bp.NewCommand(bp.SKSREG, uint8(2), fmt.Sprintf("%02X", pan.Channel())))
	ctrl.Send(bp.NewCommand(bp.SKSREG, uint8(3), fmt.Sprintf("%04X", pan.PanId())))

	k := ctrl.Send(bp.NewCommand(bp.SKLL64, uint8(3), pan.Addr()))
	addr := net.ParseIP(k.(bp.Result).Result())

	var f echonet.Frame
	ctrl.Send(bp.NewCommand(bp.SKJOIN, addr),
		func(e bp.Event) bool {
			return e.Type() == bp.EVENT && (e.(bp.EventEvent).Num() == 0x24 || e.(bp.EventEvent).Num() == 0x25)
		},
		func(e bp.Event) bool {
			if e.Type() == bp.ERXUDP {
				f = echonet.NewFrame().Decode(e.(bp.EventRxUDP).Data())
				return f.Esv() == echonet.ESV_INF
			}
			return false
		})

	var index uint8
	if f.Esv() == echonet.ESV_INF {
		index = func() uint8 {
			for _, p := range f.Properties() {
				if p.Epc() == echonet.EPC_0EF0_INSTANCE_LIST_NOTIFICATION {
					r := bytes.NewBuffer(p.Edt())
					l, _ := r.ReadByte()
					for i := 0; i < int(l); i++ {
						var cls echonet.Class
						var idx uint8
						binary.Read(r, binary.BigEndian, &cls)
						binary.Read(r, binary.BigEndian, &idx)
						if cls == echonet.CLASS_SMART_EE_METER {
							return idx
						}
					}
				}
			}
			return 0
		}()
	}

	if index <= 0 {
		log.Critical("No indexes found.")
		return
	}

	req := echonet.NewFrame()
	req.SetSeoj(echonet.CLASS_CONTROLLER, 1)
	req.SetDeoj(echonet.CLASS_SMART_EE_METER, index)
	req.SetEsv(echonet.ESV_GET)
	req.SetOpc(1)
	p := echonet.NewProperty()
	p.SetEpc(echonet.EPC_0288_UNIT_FOR_CM_AMTS_OF_EE)
	p.SetPdc(0)
	req.SetProperties([]echonet.Property{p})

	var unit float32
	ctrl.Send(bp.NewCommand(bp.SKSENDTO, uint8(1), addr, uint16(3610), uint8(1), req.Encode(getTranId())),
		func(e bp.Event) bool {
			if e.Type() == bp.ERXUDP {
				f := echonet.NewFrame().Decode(e.(bp.EventRxUDP).Data())
				seoj, idx := f.Seoj()
				if seoj == echonet.CLASS_SMART_EE_METER && idx == index && f.Esv() == echonet.ESV_GET_RES {
					p := f.Properties()[0]
					switch p.Edt()[0] {
					case 0x00: //1kWh
						unit = float32(1.0)
					case 0x01: // 0.1kWh
						unit = float32(1.e-1)
					case 0x02: // 0.01kWh
						unit = float32(1.e-2)
					case 0x03: // 0.001kWh
						unit = float32(1.e-3)
					case 0x04: // 0.0001kWh
						unit = float32(1.e-4)
					case 0x0A: // 10kWh
						unit = float32(1.e+1)
					case 0x0B: // 100kWh
						unit = float32(1.e+2)
					case 0x0C: // 1000kWh
						unit = float32(1.e+3)
					case 0x0D: // 10000kWh
						unit = float32(1.e+4)
					}
					return true
				}
			}
			return false
		})

	cli, _ := client.NewUDPClient(client.UDPConfig{Addr: fmt.Sprintf("%s:%d", conf.Database.Host, conf.Database.Port)})

	ctrl.RegisterHandler(bp.ERXUDP,
		func(e bp.Event) {
			f := echonet.NewFrame().Decode(e.(bp.EventRxUDP).Data())
			seoj, idx := f.Seoj()
			if seoj == echonet.CLASS_SMART_EE_METER && idx == index && f.Esv() == echonet.ESV_GET_RES {
				for _, p := range f.Properties() {
					bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
						Database:  "wattmeter",
						Precision: "s",
					})

					switch p.Epc() {
					case echonet.EPC_0288_CM_AMTS_OF_EE_AT_FT_NDIR:
						b := p.Edt()
						loc, _ := time.LoadLocation("Local")
						t := time.Date(
							int(binary.BigEndian.Uint16(b[0:2])), time.Month(b[2]), int(b[3]),
							int(b[4]), int(b[5]), int(b[6]), 0, loc)

						tags := map[string]string{}
						fields := map[string]interface{}{"watthour": unit * float32(binary.BigEndian.Uint32(b[7:11]))}
						pt, err := client.NewPoint("WattHour", tags, fields, t)
						if err != nil {
							log.Error(err.Error())
						} else {
							bp.AddPoint(pt)
						}
					case echonet.EPC_0288_INST_EE:
						tags := map[string]string{}
						fields := map[string]interface{}{"watt": binary.BigEndian.Uint32(p.Edt())}
						pt, err := client.NewPoint("Watt", tags, fields)
						if err != nil {
							log.Error(err.Error())
						} else {
							bp.AddPoint(pt)
						}
					}
					cli.Write(bp)
				}
			}
		})

	cr := cron.New()
	cr.AddFunc("5 */10 * * * *", func() {
		req := echonet.NewFrame()
		req.SetSeoj(echonet.CLASS_CONTROLLER, 1)
		req.SetDeoj(echonet.CLASS_SMART_EE_METER, index)
		req.SetEsv(echonet.ESV_GET)
		req.SetOpc(1)
		p := echonet.NewProperty()
		p.SetEpc(echonet.EPC_0288_CM_AMTS_OF_EE_AT_FT_NDIR)
		p.SetPdc(0)
		req.SetProperties([]echonet.Property{p})

		ctrl.Send(bp.NewCommand(bp.SKSENDTO, uint8(1), addr, uint16(3610), uint8(1), req.Encode(getTranId())))
	})

	cr.AddFunc("*/10 * * * * *", func() {
		req := echonet.NewFrame()
		req.SetSeoj(echonet.CLASS_CONTROLLER, 1)
		req.SetDeoj(echonet.CLASS_SMART_EE_METER, index)
		req.SetEsv(echonet.ESV_GET)
		req.SetOpc(1)
		p := echonet.NewProperty()
		p.SetEpc(echonet.EPC_0288_INST_EE)
		p.SetPdc(0)
		req.SetProperties([]echonet.Property{p})

		ctrl.Send(bp.NewCommand(bp.SKSENDTO, uint8(1), addr, uint16(3610), uint8(1), req.Encode(getTranId())))
	})

	cr.Start()

	select {}
}

func configLogger(level string) {
	defer log.Flush()

	loglv, f := log.LogLevelFromString(level)
	if !f {
		loglv = log.InfoLvl
	}
	writer, _ := log.NewBufferedWriter(os.Stderr, 128, 500)
	formatter, _ := log.NewFormatter("%Date %Time [%LEV]: %Msg%n")
	root, _ := log.NewSplitDispatcher(formatter, []interface{}{writer})
	constraints, _ := log.NewMinMaxConstraints(loglv, log.CriticalLvl)
	exceptions := []*log.LogLevelException{}

	logger := log.NewAsyncLoopLogger(log.NewLoggerConfig(constraints, exceptions, root))
	log.ReplaceLogger(logger)
}

func getTTYPath() (string, error) {
	u := udev.Udev{}
	var dev *udev.Device
	var err error

	dev, err = findDevice(&u, func(enum *udev.Enumerate) {
		enum.AddMatchSubsystem("usb")
		enum.AddMatchSysattr("interface", "FT232R USB UART")
		enum.AddMatchIsInitialized()
	})
	if err != nil {
		return "", err
	}

	dev, err = findDevice(&u, func(enum *udev.Enumerate) {
		enum.AddMatchSubsystem("tty")
		enum.AddMatchParent(dev)
	})
	if err != nil {
		return "", err
	}

	return dev.Devnode(), nil
}

func findDevice(u *udev.Udev, filter func(*udev.Enumerate)) (*udev.Device, error) {
	enum := u.NewEnumerate()
	filter(enum)
	devices, err := enum.Devices()

	if len(devices) <= 0 {
		return nil, errors.New("No devices found.")
	}
	return devices[0], err
}
