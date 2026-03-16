package ag_netty

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- test helpers ----------

// clientEchoHandler — client-side handler that completes Future on read, and writes directly on write
type clientEchoHandler struct{}

func (h *clientEchoHandler) HandleActive(ctx *HandlerContext)           {}
func (h *clientEchoHandler) HandleInactive(ctx *HandlerContext)         {}
func (h *clientEchoHandler) HandleError(ctx *HandlerContext, err error) {}
func (h *clientEchoHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	ctx.Channel().WriteDirect(data)
}
func (h *clientEchoHandler) HandleRead(ctx *HandlerContext, data []byte) {
	ctx.Channel().Future().Complete(string(data))
}

// collectHandler — collects received data for assertion
type collectHandler struct {
	mu   sync.Mutex
	data [][]byte
	done chan struct{} // closed on first read
	once sync.Once
}

func newCollectHandler() *collectHandler {
	return &collectHandler{done: make(chan struct{})}
}

func (h *collectHandler) HandleActive(ctx *HandlerContext)           {}
func (h *collectHandler) HandleInactive(ctx *HandlerContext)         {}
func (h *collectHandler) HandleError(ctx *HandlerContext, err error) {}
func (h *collectHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	ctx.Channel().WriteDirect(data)
}
func (h *collectHandler) HandleRead(ctx *HandlerContext, data []byte) {
	h.mu.Lock()
	h.data = append(h.data, append([]byte(nil), data...))
	h.mu.Unlock()
	h.once.Do(func() { close(h.done) })
}

func (h *collectHandler) waitFirst(timeout time.Duration) bool {
	select {
	case <-h.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (h *collectHandler) getAll() [][]byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([][]byte, len(h.data))
	copy(out, h.data)
	return out
}

// startEchoServer starts a plain TCP echo server on a random port and returns it
func startEchoServer(t *testing.T) *Server {
	t.Helper()
	srv, err := NewServer("127.0.0.1:0", func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &EchoHandler{})
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv.Start()
	return srv
}

func serverAddr(srv *Server) string {
	return srv.Addr().String()
}

// ---------- tests ----------

func TestBasicEcho(t *testing.T) {
	srv := startEchoServer(t)
	defer srv.Shutdown()

	addr := serverAddr(srv)

	collector := newCollectHandler()
	client := NewClient(addr, 2*time.Second, 2*time.Second, 2*time.Second, 0, func(ch *Channel) {
		ch.Pipeline.AddLast("collect", collector)
	})
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	client.Send([]byte("hello"))

	if !collector.waitFirst(3 * time.Second) {
		t.Fatal("timeout waiting for echo response")
	}

	got := collector.getAll()
	if len(got) == 0 {
		t.Fatal("no data received")
	}
	if string(got[0]) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(got[0]))
	}
}

func TestSendAndGet(t *testing.T) {
	srv := startEchoServer(t)
	defer srv.Shutdown()

	addr := serverAddr(srv)

	client := NewClient(addr, 2*time.Second, 2*time.Second, 2*time.Second, 0, func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &clientEchoHandler{})
	})
	defer client.Close()

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	result, err := client.SendAndGet([]byte("ping"))
	if err != nil {
		t.Fatalf("SendAndGet: %v", err)
	}

	if result != "ping" {
		t.Fatalf("expected 'ping', got %v", result)
	}
}

func TestMultipleConnections(t *testing.T) {
	srv := startEchoServer(t)
	defer srv.Shutdown()

	addr := serverAddr(srv)
	const N = 10

	var wg sync.WaitGroup
	errors := make(chan error, N)

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			c := NewClient(addr, 2*time.Second, 2*time.Second, 2*time.Second, 0, func(ch *Channel) {
				ch.Pipeline.AddLast("echo", &clientEchoHandler{})
			})
			defer c.Close()

			if err := c.Connect(); err != nil {
				errors <- fmt.Errorf("client %d connect: %w", id, err)
				return
			}

			msg := fmt.Sprintf("hello-%d", id)
			result, err := c.SendAndGet([]byte(msg))
			if err != nil {
				errors <- fmt.Errorf("client %d sendandget: %w", id, err)
				return
			}
			if result != msg {
				errors <- fmt.Errorf("client %d: expected %q, got %v", id, msg, result)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestFutureDoubleComplete(t *testing.T) {
	f := NewFuture()

	// First complete should work
	f.Complete("ok")

	// Second complete should not panic
	f.Complete("should-be-ignored")

	result := f.Get()
	if result != "ok" {
		t.Fatalf("expected 'ok', got %v", result)
	}
}

func TestFutureGetWithTimeout(t *testing.T) {
	f := NewFuture()

	// Should timeout
	_, err := f.GetWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	// Complete after timeout
	f.Complete("late")

	// Should return immediately now
	result, err := f.GetWithTimeout(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "late" {
		t.Fatalf("expected 'late', got %v", result)
	}
}

func TestGracefulShutdown(t *testing.T) {
	srv := startEchoServer(t)
	addr := serverAddr(srv)

	// Connect a client
	client := NewClient(addr, 2*time.Second, 2*time.Second, 2*time.Second, 0, func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &clientEchoHandler{})
	})

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Verify it works
	result, err := client.SendAndGet([]byte("before-shutdown"))
	if err != nil {
		t.Fatalf("SendAndGet: %v", err)
	}
	if result != "before-shutdown" {
		t.Fatalf("expected 'before-shutdown', got %v", result)
	}

	// Shutdown server — should not hang or panic
	done := make(chan struct{})
	go func() {
		srv.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown hung for >5s")
	}

	client.Close()
}

func TestEventLoopPostAfterShutdown(t *testing.T) {
	el := NewEventLoop()
	el.Shutdown()

	// Post after shutdown should not block
	done := make(chan struct{})
	go func() {
		el.Post(func() { t.Log("should not execute") })
		close(done)
	}()

	select {
	case <-done:
		// OK — Post returned
	case <-time.After(2 * time.Second):
		t.Fatal("Post blocked after Shutdown")
	}
}

func TestEventLoopTaskDrain(t *testing.T) {
	el := NewEventLoop()

	var count atomic.Int32

	// Post many tasks
	for i := 0; i < 100; i++ {
		el.Post(func() {
			count.Add(1)
		})
	}

	// Give the event loop time to process
	time.Sleep(100 * time.Millisecond)

	el.Shutdown()

	if c := count.Load(); c != 100 {
		t.Fatalf("expected 100 tasks executed, got %d", c)
	}
}

func TestChannelActiveFlagRace(t *testing.T) {
	// Create a channel with a real connection to test race safety
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Accept in background
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewNetConnAdapter(conn)
	el := NewEventLoop()
	defer el.Shutdown()

	ch := NewChannel(adapter, el)
	ch.Pipeline.AddLast("echo", &EchoHandler{})

	// Concurrent reads of IsActive + Close should not race
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ch.IsActive()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ch.Close()
	}()
	wg.Wait()

	if ch.IsActive() {
		t.Fatal("channel should be inactive after Close")
	}
}

func TestServerAddrRandomPort(t *testing.T) {
	srv, err := NewServer("127.0.0.1:0", func(ch *Channel) {
		ch.Pipeline.AddLast("echo", &EchoHandler{})
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	srv.Start()
	defer srv.Shutdown()

	addr := srv.Addr()
	if addr == nil {
		t.Fatal("expected non-nil Addr")
	}

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected *net.TCPAddr, got %T", addr)
	}
	if tcpAddr.Port == 0 {
		t.Fatal("expected non-zero port")
	}
	t.Logf("server listening on %s", addr)
}

func TestPipelineHandlerOrder(t *testing.T) {
	// Verify handler chain execution order
	var order []string
	var mu sync.Mutex

	makeHandler := func(name string) ChannelHandler {
		return &orderHandler{
			name:  name,
			order: &order,
			mu:    &mu,
		}
	}

	srv, err := NewServer("127.0.0.1:0", func(ch *Channel) {
		ch.Pipeline.AddLast("first", makeHandler("first"))
		ch.Pipeline.AddLast("second", makeHandler("second"))
	})
	if err != nil {
		t.Fatal(err)
	}
	srv.Start()
	defer srv.Shutdown()

	// Connect and send data
	conn, err := net.Dial("tcp", srv.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte("test"))
	time.Sleep(200 * time.Millisecond)
	conn.Close()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Read events should fire first→second (inbound direction)
	foundRead := false
	for i, entry := range order {
		if entry == "first:read" {
			if i+1 < len(order) && order[i+1] == "second:read" {
				foundRead = true
			}
			break
		}
	}
	if !foundRead {
		t.Errorf("expected read order first→second, got %v", order)
	}
}

type orderHandler struct {
	name  string
	order *[]string
	mu    *sync.Mutex
}

func (h *orderHandler) HandleActive(ctx *HandlerContext) {
	h.record("active")
}
func (h *orderHandler) HandleInactive(ctx *HandlerContext) {
	h.record("inactive")
}
func (h *orderHandler) HandleRead(ctx *HandlerContext, data []byte) {
	h.record("read")
}
func (h *orderHandler) HandleWrite(ctx *HandlerContext, data []byte) {
	h.record("write")
	ctx.Channel().WriteDirect(data)
}
func (h *orderHandler) HandleError(ctx *HandlerContext, err error) {
	h.record("error")
}
func (h *orderHandler) record(event string) {
	h.mu.Lock()
	*h.order = append(*h.order, h.name+":"+event)
	h.mu.Unlock()
}
