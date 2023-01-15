package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServeHTTP(t *testing.T) {
	t.Run("when context done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var port int32 = 46666
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			assert.NoError(t, ServeHTTP(ctx, port))
			wg.Done()
		}()
		time.Sleep(1 * time.Second)

		resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:%d/ping", port))
		assert.NoError(t, err)
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(data))

		ps := []string{
			"foo",
			"bar",
			time.Now().String(),
		}
		for i := range ps {
			resp, err := http.Post(
				fmt.Sprintf("http://0.0.0.0:%d/echo", port),
				"text/plain",
				bytes.NewBufferString(ps[i]),
			)
			assert.NoError(t, err)
			data, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, ps[i], string(data))
		}
		cancel()
		wg.Wait()
	})
}
