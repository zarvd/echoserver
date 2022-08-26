package main

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parsePortsString(t *testing.T) {
	t.Run("parse empty string", func(t *testing.T) {
		ports, err := parsePortsString("")
		assert.NoError(t, err)
		assert.Empty(t, ports)
	})

	t.Run("parse invalid string", func(t *testing.T) {
		ports, err := parsePortsString("foo")
		assert.Error(t, err)
		assert.Empty(t, ports)
	})

	t.Run("parse 1 port", func(t *testing.T) {
		ports, err := parsePortsString("123")
		assert.NoError(t, err)
		assert.Len(t, ports, 1)
		assert.Equal(t, int32(123), ports[0])
	})

	t.Run("parse multiple ports", func(t *testing.T) {
		ports, err := parsePortsString("123,321,123,555")
		assert.NoError(t, err)
		assert.Len(t, ports, 3)
		assert.Contains(t, ports, int32(123))
		assert.Contains(t, ports, int32(321))
		assert.Contains(t, ports, int32(555))
	})

	t.Run("parse port range", func(t *testing.T) {
		ports, err := parsePortsString("123,500-599,321,555,1000-1999")
		assert.NoError(t, err)
		assert.Len(t, ports, 1102)
		assert.Contains(t, ports, int32(123))
		assert.Contains(t, ports, int32(321))
		for i := 500; i < 600; i++ {
			assert.Contains(t, ports, int32(i))
		}
		for i := 1000; i < 2000; i++ {
			assert.Contains(t, ports, int32(i))
		}
	})
}

func TestServe(t *testing.T) {
	t.Run("when not enabled", func(t *testing.T) {
		ctx := context.Background()
		wg := &sync.WaitGroup{}
		Serve(ctx, wg, "UDP", false, "", ServeUDP)
		wg.Wait()
	})

	t.Run("serve udp", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		Serve(ctx, wg, "UDP", true, "43325", ServeUDP)
		cancel()
		wg.Wait()
	})
}
