package ag_netty

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func tclpBenchEchoServer(b *testing.B) (*Server, string) {
	b.Helper()
	srv, err := NewServer("127.0.0.1:0", func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &EchoHandler{})
	})
	if err != nil {
		b.Fatalf("NewServer: %v", err)
	}
	srv.Start()
	return srv, srv.Addr().String()
}

func tclpBenchClient(b *testing.B, addr string) *Client {
	b.Helper()
	c := NewClient(addr, 5*time.Second, 5*time.Second, 5*time.Second, 0, func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &clientEchoHandler{})
	})
	if err := c.Connect(); err != nil {
		b.Fatalf("Connect: %v", err)
	}
	return c
}

func BenchmarkTlcpEchoRoundTrip(b *testing.B) {
	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	c := tclpBenchClient(b, addr)
	defer c.Close()

	payload := []byte("benchmark-ping")
	if _, err := c.SendAndGet(payload); err != nil {
		b.Fatalf("warmup: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := c.SendAndGet(payload); err != nil {
			b.Fatalf("iter %d: %v", i, err)
		}
	}
}

func BenchmarkTlcpEchoPayloadSize(b *testing.B) {
	sizes := []int{64, 512, 1024, 4096}

	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			c := tclpBenchClient(b, addr)
			defer c.Close()

			payload := make([]byte, size)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			if _, err := c.SendAndGet(payload); err != nil {
				b.Fatalf("warmup: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				if _, err := c.SendAndGet(payload); err != nil {
					b.Fatalf("iter %d: %v", i, err)
				}
			}
		})
	}
}

func BenchmarkTlcpEchoParallel(b *testing.B) {
	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	payload := []byte("parallel-ping")

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		c := tclpBenchClient(b, addr)
		defer c.Close()

		for pb.Next() {
			if _, err := c.SendAndGet(payload); err != nil {
				b.Errorf("SendAndGet: %v", err)
				return
			}
		}
	})
}

func BenchmarkTlcpSendFireAndForget(b *testing.B) {
	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	collector := newCollectHandler()
	c := NewClient(addr, 5*time.Second, 5*time.Second, 5*time.Second, 0, func(ch *Channel) {
		ch.Pipeline.AddLast("collect", collector)
	})
	if err := c.Connect(); err != nil {
		b.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	payload := []byte("fire-and-forget")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Send(payload)
	}

	collector.waitFirst(3 * time.Second)
}

func BenchmarkTlcpEventLoopPost(b *testing.B) {
	el := NewEventLoop()
	defer el.Shutdown()

	var counter atomic.Int64

	b.ResetTimer()
	b.ReportAllocs()

	done := make(chan struct{})
	remaining := int64(b.N)

	for i := 0; i < b.N; i++ {
		el.Post(func() {
			counter.Add(1)
			if counter.Load() == remaining {
				close(done)
			}
		})
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.Fatalf("timeout waiting for %d tasks, completed %d", b.N, counter.Load())
	}
}

func BenchmarkTlcpEventLoopPostParallel(b *testing.B) {
	el := NewEventLoop()
	defer el.Shutdown()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			el.Post(func() {})
		}
	})
}

func BenchmarkTlcpFutureCreateComplete(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		f := NewFuture()
		f.Complete("result")
		_ = f.Get()
	}
}

func BenchmarkTlcpPipelineFireRead(b *testing.B) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	go func() {
		buf := make([]byte, 65536)
		for {
			if _, err := server.Read(buf); err != nil {
				return
			}
		}
	}()

	adapter := NewNetConnAdapter(client)
	el := NewEventLoop()
	defer el.Shutdown()

	ch := NewChannel(adapter, el)
	var readCount atomic.Int64
	ch.Pipeline.AddLast("counter", &countingHandler{count: &readCount})

	payload := []byte("pipeline-data")

	b.ResetTimer()
	b.ReportAllocs()

	done := make(chan struct{})
	target := int64(b.N)

	for i := 0; i < b.N; i++ {
		el.Post(func() {
			ch.Pipeline.FireRead(payload)
			if readCount.Load() == target {
				close(done)
			}
		})
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.Fatalf("timeout: processed %d/%d", readCount.Load(), b.N)
	}
}

func BenchmarkTlcpPipelineMultiHandler(b *testing.B) {
	for _, nHandlers := range []int{1, 3, 5} {
		b.Run(fmt.Sprintf("%d_handlers", nHandlers), func(b *testing.B) {
			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			go func() {
				buf := make([]byte, 65536)
				for {
					if _, err := server.Read(buf); err != nil {
						return
					}
				}
			}()

			adapter := NewNetConnAdapter(client)
			el := NewEventLoop()
			defer el.Shutdown()

			ch := NewChannel(adapter, el)
			var readCount atomic.Int64
			for i := 0; i < nHandlers; i++ {
				ch.Pipeline.AddLast(fmt.Sprintf("handler-%d", i), &passThroughHandler{count: &readCount})
			}

			payload := []byte("pipeline-multi")

			b.ResetTimer()
			b.ReportAllocs()

			done := make(chan struct{})
			target := int64(b.N) * int64(nHandlers)

			for i := 0; i < b.N; i++ {
				el.Post(func() {
					ch.Pipeline.FireRead(payload)
					if readCount.Load() == target {
						close(done)
					}
				})
			}

			select {
			case <-done:
			case <-time.After(30 * time.Second):
				b.Fatalf("timeout: processed %d/%d", readCount.Load(), target)
			}
		})
	}
}

func BenchmarkTlcpWriteDirect(b *testing.B) {
	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	adapter := NewNetConnAdapter(conn)
	el := NewEventLoop()
	defer el.Shutdown()

	ch := NewChannel(adapter, el)
	ch.Pipeline.AddLast("echo", &EchoHandler{})

	payload := []byte("direct-write-bench")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := ch.WriteDirect(payload); err != nil {
			b.Fatalf("WriteDirect: %v", err)
		}
	}
}

func BenchmarkTlcpConnectDisconnect(b *testing.B) {
	srv, addr := tclpBenchEchoServer(b)
	defer srv.Shutdown()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c := NewClient(addr, 2*time.Second, 2*time.Second, 2*time.Second, 0, func(ch *Channel) {
			ch.Pipeline.AddLast("echo", &clientEchoHandler{})
		})
		if err := c.Connect(); err != nil {
			b.Fatalf("Connect: %v", err)
		}
		c.Close()
	}
}

func BenchmarkTlcpConcurrentConnections(b *testing.B) {
	for _, numConns := range []int{1, 5, 10, 50} {
		b.Run(fmt.Sprintf("%d_conns", numConns), func(b *testing.B) {
			srv, addr := tclpBenchEchoServer(b)
			defer srv.Shutdown()

			clients := make([]*Client, numConns)
			for i := 0; i < numConns; i++ {
				clients[i] = tclpBenchClient(b, addr)
			}
			defer func() {
				for _, c := range clients {
					c.Close()
				}
			}()

			payload := []byte("concurrent-bench")

			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			opsPerClient := b.N / numConns
			if opsPerClient == 0 {
				opsPerClient = 1
			}

			for _, c := range clients {
				wg.Add(1)
				go func(client *Client) {
					defer wg.Done()
					for j := 0; j < opsPerClient; j++ {
						if _, err := client.SendAndGet(payload); err != nil {
							b.Errorf("SendAndGet: %v", err)
							return
						}
					}
				}(c)
			}
			wg.Wait()
		})
	}
}
