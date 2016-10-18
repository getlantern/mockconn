package conntest

import (
	"bytes"
	"net"
	"sync"
	"time"
)

// Dialer is a test dialer that provides a net.Dial and net.DialTimeout
// equivalent backed by an in-memory data structure, and that provides access to
// the received data via the Received() method.
type Dialer interface {
	Dial(network, addr string) (net.Conn, error)
	DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error)
	Received() []byte
}

// NewDialer constructs a new Dialer that responds with the given canned
// responseData.
func NewDialer(responseData []byte) Dialer {
	return &dialer{
		responseData: responseData,
		received:     &bytes.Buffer{},
	}
}

type dialer struct {
	responseData []byte
	received     *bytes.Buffer
	mx           sync.RWMutex
}

func (d *dialer) Dial(network, addr string) (net.Conn, error) {
	return &conn{d, bytes.NewBuffer(d.responseData)}, nil
}

func (d *dialer) DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	return d.Dial(network, addr)
}

func (d *dialer) Received() []byte {
	d.mx.RLock()
	defer d.mx.RUnlock()
	return d.received.Bytes()
}

type conn struct {
	d            *dialer
	responseData *bytes.Buffer
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.responseData.Read(b)
}

func (c *conn) Write(b []byte) (n int, err error) {
	c.d.mx.Lock()
	defer c.d.mx.Unlock()
	return c.d.received.Write(b)
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	return nil
}

func (c *conn) RemoteAddr() net.Addr {
	return nil
}

func (c *conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}
