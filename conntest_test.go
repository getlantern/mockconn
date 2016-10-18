package conntest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDial(t *testing.T) {
	d := NewDialer([]byte("Response"))
	conn, err := d.Dial("tcp", "doesn't matter")
	if !assert.NoError(t, err) {
		return
	}
	n, err := conn.Write([]byte("Request"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 7, n)
	received := d.Received()
	assert.Equal(t, string(received), "Request")
}
