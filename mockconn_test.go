package mockconn

import (
	"errors"
	"testing"

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
}

func TestFailingDialer(t *testing.T) {
	err := errors.New("Test error")
	d := FailingDialer(err)
	_, actualErr := d.Dial("tcp", "doesn't matter")
	assert.Equal(t, err, actualErr)
}
