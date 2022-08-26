package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServeUDP(t *testing.T) {
	t.Run("when context done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var port int32 = 45555
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			assert.NoError(t, ServeUDP(ctx, port))
			wg.Done()
		}()

		time.Sleep(1 * time.Second)

		ps := []string{
			"foo\n",
			"bar\n",
			time.Now().String() + "\n",
		}
		for i := range ps {
			conn, err := net.Dial("udp", fmt.Sprintf("0.0.0.0:%d", port))
			assert.NoError(t, err)
			assert.NoError(t, conn.SetDeadline(time.Now().Add(2*time.Second)))
			reader := bufio.NewReader(conn)
			n, err := conn.Write([]byte(ps[i]))
			assert.NoError(t, err)
			assert.Equal(t, len(ps[i]), n)

			b, err := reader.ReadBytes(byte('\n'))
			assert.NoError(t, err)
			assert.Equal(t, ps[i], string(b))
		}
		cancel()
		wg.Wait()
	})
}
