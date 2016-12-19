package mockconn

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"
)

// Dialer is a test dialer that provides a net.Dial and net.DialTimeout
// equivalent backed by an in-memory data structure, and that provides access to
// the received data via the Received() method.
type Dialer interface {
	// Like net.Dial
	Dial(network, addr string) (net.Conn, error)

	// Like net.DialTimeout
	DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error)

	// Gets the last dialed address
	LastDialed() string

	// Gets all received data
	Received() []byte

	// Returns true if all dialed connections are closed
	AllClosed() bool
}

// SucceedingDialer constructs a new Dialer that responds with the given canned
// responseData.
func SucceedingDialer(responseData []byte) Dialer {
	var mx sync.RWMutex
	return &dialer{
		responseData: responseData,
		received:     &bytes.Buffer{},
		mx:           &mx,
	}
}

// FailingDialer constructs a new Dialer that fails to dial with the given
// error.
func FailingDialer(dialError error) Dialer {
	var mx sync.RWMutex
	return &dialer{
		dialError: dialError,
		mx:        &mx,
	}
}

type dialer struct {
	dialError    error
	responseData []byte
	lastDialed   string
	numOpen      int
	received     *bytes.Buffer
	mx           *sync.RWMutex
}

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	d.lastDialed = addr
	if d.dialError != nil {
		return nil, d.dialError
	}
	d.numOpen++
	return &Conn{
		responseData: bytes.NewBuffer(d.responseData),
		received:     d.received,
		mx:           d.mx,
		onClose: func() {
			d.numOpen--
		},
	}, nil
}

func (d *dialer) DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	return d.Dial(network, addr)
}

func (d *dialer) LastDialed() string {
	d.mx.RLock()
	defer d.mx.RUnlock()
	return d.lastDialed
}

func (d *dialer) Received() []byte {
	d.mx.RLock()
	defer d.mx.RUnlock()
	return d.received.Bytes()
}

func (d *dialer) AllClosed() bool {
	d.mx.RLock()
	defer d.mx.RUnlock()
	return d.numOpen == 0
}

func New(received *bytes.Buffer, responseData io.Reader) *Conn {
	var mx sync.RWMutex
	return &Conn{received: received, responseData: responseData, mx: &mx}
}

type Conn struct {
	responseData io.Reader
	received     *bytes.Buffer
	closed       bool
	onClose      func()
	mx           *sync.RWMutex
}

func (c *Conn) Read(b []byte) (n int, err error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	return c.responseData.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.received.Write(b)
}

func (c *Conn) Close() error {
	c.mx.Lock()
	defer c.mx.Unlock()
	if !c.closed {
		c.closed = true
		if c.onClose != nil {
			c.onClose()
		}
	}
	return nil
}

func (c *Conn) Closed() bool {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.closed
}

func (c *Conn) LocalAddr() net.Addr {
	return nil
}

func (c *Conn) RemoteAddr() net.Addr {
	return nil
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
