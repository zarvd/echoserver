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

func TestHandleTCPConnection(t *testing.T) {
	t.Run("should stop if no incoming bytes", func(t *testing.T) {
		tx, rx := net.Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			HandleTCPConnection(rx)
			wg.Done()
		}()
		assert.NoError(t, tx.SetReadDeadline(time.Now().Add(1*time.Second)))
		b := make([]byte, 4096)
		_, err := tx.Read(b)
		assert.Error(t, err)
		assert.NoError(t, tx.Close())
		wg.Wait()
	})

	t.Run("should echo the same value", func(t *testing.T) {
		tx, rx := net.Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			HandleTCPConnection(rx)
			wg.Done()
		}()
		assert.NoError(t, tx.SetDeadline(time.Now().Add(2*time.Second)))

		ps := []string{
			"foo\n",
			"bar\n",
			time.Now().String() + "\n",
		}
		reader := bufio.NewReader(tx)
		for i := range ps {
			n, err := tx.Write([]byte(ps[i]))
			assert.NoError(t, err)
			assert.Equal(t, len(ps[i]), n)

			b, err := reader.ReadBytes(byte('\n'))
			assert.NoError(t, err)
			assert.Equal(t, ps[i], string(b))
		}

		b := make([]byte, 4096)
		_, err := tx.Read(b)
		assert.Error(t, err)
		assert.NoError(t, tx.Close())
		wg.Wait()
	})
}

func TestServeTCP(t *testing.T) {
	t.Run("when context done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var port int32 = 45555
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			assert.NoError(t, ServeTCP(ctx, port))
			wg.Done()
		}()

		time.Sleep(1 * time.Second)

		conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		assert.NoError(t, err)
		assert.NoError(t, conn.SetDeadline(time.Now().Add(2*time.Second)))

		ps := []string{
			"foo\n",
			"bar\n",
			time.Now().String() + "\n",
		}
		reader := bufio.NewReader(conn)
		for i := range ps {
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
