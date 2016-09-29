package bp35a1

import (
	"bufio"
	"github.com/tarm/serial"
	"io"
	"strings"
	"sync"
	"time"

	log "github.com/cihub/seelog"
)

type handler func(Event)
type condition func(Event) bool

type Controller interface {
	Send(Command, ...condition) Response
	RegisterHandler(ev, ...handler)
}

type controller struct {
	handlers map[ev][]handler
	watchers map[chan<- Event]func()
	mutex    *sync.Mutex
	send     chan Command
	recv     chan Event
	resp     chan Response
}

func NewController(tty string) Controller {
	ser, err := serial.OpenPort(&serial.Config{Name: tty, Baud: 115200})
	if err != nil {
		log.Critical(err)
	}

	c := &controller{
		handlers: make(map[ev][]handler),
		watchers: make(map[chan<- Event]func()),
		mutex:    new(sync.Mutex),
		send:     make(chan Command),
		recv:     make(chan Event),
		resp:     make(chan Response)}

	resp := make(chan interface{})
	go c.reciever(ser, resp)
	go c.sender(ser, resp)
	go c.processEvent()

	return c
}

func (c *controller) Send(cmd Command, cond ...condition) Response {
	var wg sync.WaitGroup
	for _, cn := range cond {
		wg.Add(1)
		w := make(chan Event)
		f := func() {
			defer wg.Done()
			defer close(w)
			defer delete(c.watchers, w)

			for {
				select {
				case e := <-w:
					if cn(e) {
						return
					}
				case <-time.After(time.Second * 10):
					return
				}
			}
		}
		c.addWatcher(w, f)
		go f()
	}

	c.send <- cmd
	r := <-c.resp

	wg.Wait()
	return r
}

func (c *controller) RegisterHandler(e ev, hdr ...handler) {
	if _, ok := c.handlers[e]; !ok {
		c.handlers[e] = []handler{}
	}

	c.handlers[e] = append(c.handlers[e], hdr...)
}

func (c *controller) addWatcher(w chan Event, f func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.watchers[w] = f
}

func (c *controller) sender(wt io.Writer, resp <-chan interface{}) {
	result := make(chan bool)
	for c := range c.send {
		go func() {
			select {
			case <-resp:
				result <- true
			case <-time.After(time.Second * 2):
				result <- false
			}
		}()
		_, err := wt.Write(append(ToBytes(c), []byte("\r\n")...))

		<-result
		if err != nil {
			log.Critical(err)
		}
	}
}

func (c *controller) reciever(rd io.Reader, resp chan<- interface{}) {
	var m MultiLine
	var ln []string

	f := func() {
		if m != nil {
			m.Parse(ln)
			c.recv <- m.(Event)
			m = nil
		}
	}

	s := bufio.NewScanner(rd)
	for s.Scan() {
		log.Debug(s.Text())
		data := s.Text()
		switch {
		case strings.HasPrefix(data, "SK"): // Ignore echo back
			f()
		case strings.HasPrefix(data, "E"):
			f()

			e := newEvent(data)
			m, _ = e.(MultiLine)
			if m != nil {
				ln = []string{}
			} else {
				c.recv <- e
			}
		case strings.HasPrefix(data, "OK"):
			f()

			c.resp <- &response{t: OK}
			resp <- true
		case strings.HasPrefix(data, "FAIL"):
			f()

			r := strings.Split(data, " ")
			c.resp <- &response_fail{response: &response{t: FAIL}, code: string(r[1])}
			resp <- false
		default:
			if m != nil {
				ln = append(ln, data)
			} else if len(data) > 0 {
				c.resp <- &response_result{response: &response{t: RESULT}, result: string(data)}
				resp <- true
			}
		}
	}
}

func (c *controller) processEvent() {
	for {
		e := <-c.recv
		if hdr, ok := c.handlers[e.Type()]; ok {
			for _, h := range hdr {
				h(e)
			}
		}

		for w, _ := range c.watchers {
			w <- e
		}
	}
}
