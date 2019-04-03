package mockconn

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSucceedingDialer(t *testing.T) {
	d := SucceedingDialer([]byte("Response"))
	conn, err := d.Dial("tcp", "doesn't matter")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "doesn't matter", d.LastDialed())
	n, err := conn.Write([]byte("Request"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 7, n)
	received := d.Received()
	assert.Equal(t, string(received), "Request")
	assert.False(t, conn.(*Conn).Closed())
	assert.False(t, d.AllClosed())
	var buf [10]byte
	n, err = conn.Read(buf[:])
	assert.Equal(t, 8, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Response"), buf[:8])
	n, err = conn.Read(buf[:])
	assert.Equal(t, io.EOF, err, "read again from the connection should get EOF")
	assert.Equal(t, 0, n)

	conn.Close()
	assert.True(t, conn.(*Conn).Closed())
	assert.True(t, d.AllClosed())
}

func TestFailingDialer(t *testing.T) {
	err := errors.New("Test error")
	d := FailingDialer(err)
	_, actualErr := d.Dial("tcp", "doesn't matter")
	assert.Equal(t, err, actualErr)
}

func TestSlowDialer(t *testing.T) {
	delay := 100 * time.Millisecond
	d := SlowDialer(SucceedingDialer([]byte{}), delay)
	start := time.Now()
	_, err := d.Dial("tcp", "doesn't matter")
	if !assert.NoError(t, err) {
		return
	}
	assert.InDelta(t, delay.Nanoseconds(), time.Since(start).Nanoseconds(), float64(10*time.Millisecond))

	expectedError := errors.New("Test error")
	d = SlowDialer(FailingDialer(expectedError), delay)
	start = time.Now()
	_, err = d.Dial("tcp", "doesn't matter")
	if !assert.Equal(t, expectedError, err) {
		return
	}
	assert.InDelta(t, delay.Nanoseconds(), time.Since(start).Nanoseconds(), float64(10*time.Millisecond))
}

func TestSlowResponder(t *testing.T) {
	delay := 50 * time.Millisecond
	d := SlowResponder(SucceedingDialer([]byte("Response")), delay)
	conn, err := d.Dial("tcp", "doesn't matter")
	if !assert.NoError(t, err) {
		return
	}
	var buf [10]byte
	start := time.Now()
	n, err := conn.Read(buf[:])
	assert.Equal(t, 8, n)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Response"), buf[:8])
	assert.InDelta(t, delay.Nanoseconds(), time.Since(start).Nanoseconds(), float64(10*time.Millisecond))
	conn.Close()
}

func TestAutoClose(t *testing.T) {
	d := AutoClose(SucceedingDialer([]byte("Response")))
	conn, err := d.Dial("tcp", "doesn't matter")
	if !assert.NoError(t, err) {
		return
	}
	_, err = conn.Write([]byte("Request"))
	if !assert.NoError(t, err) {
		return
	}
	var buf [10]byte
	_, _ = conn.Read(buf[:])
	assert.Equal(t, []byte("Response"), buf[:8])
	t.Log(d.(*dialer).numOpen)
	assert.True(t, d.AllClosed(), "connection should be closed automatically")
}
