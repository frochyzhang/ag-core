package ag_netty

import (
	"github.com/cloudwego/netpoll"
	"log/slog"
	"time"
)

// Dial 连接到服务器
func Dial(
	addr string,
	connTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	looper EventLooper,
) (*Channel, error) {
	conn, err := netpoll.DialConnection("tcp", addr, connTimeout)
	if err != nil {
		return nil, err
	}

	// 创建通道
	channel := NewChannel(conn, looper)

	// 初始化Pipeline
	if clientLooper, ok := looper.(*ClientEventLoop); ok {
		if clientLooper.initFunc != nil {
			clientLooper.initFunc(channel)
		}
	}

	// 触发激活事件
	looper.Post(func() {
		channel.Pipeline.FireActive()
	})

	// 设置读超时
	conn.SetReadTimeout(readTimeout)
	// 设置写超时
	conn.SetWriteTimeout(writeTimeout)
	// 设置空闲超时
	conn.SetIdleTimeout(idleTimeout)

	// 启动读循环
	go readLoop(conn, channel, looper)

	slog.Info("Connected to server", "addr", addr)
	return channel, nil
}

// readLoop 读取数据循环
func readLoop(
	conn netpoll.Connection,
	channel *Channel,
	looper EventLooper,
) {
	reader := conn.Reader()
	for {
		if looper.IsShutdown() {
			return
		}

		// 检查可读数据
		n := reader.Len()

		if n == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// 读取数据
		data, err := reader.ReadBinary(n)
		if err != nil {
			looper.Post(func() {
				channel.Pipeline.FireError(err)
			})
			channel.Close()
			return
		}

		// 触发读事件
		looper.Post(func() {
			channel.Pipeline.FireRead(data)
		})
	}
}
