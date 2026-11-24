// Command loadtest measures end-to-end produce throughput and latency.
//
//	go run ./cmd/loadtest [size] [count] [mode] [n]
//
// mode is seq (default, one request in flight), pipe (n requests in flight on
// one connection), or multi (n connections, one request in flight each).
//
//	go run ./cmd/loadtest 1024 20000 seq
//	go run ./cmd/loadtest 1024 20000 pipe 64
//	go run ./cmd/loadtest 1024 20000 multi 64
package main

import (
	"bufio"
	"fmt"
	"kafkaG/internal/protocol"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	addr  = "127.0.0.1:9092"
	topic = "bench"
)

type client struct {
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
}

func dial() *client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	return &client{conn, bufio.NewWriterSize(conn, 64<<10), bufio.NewReader(conn)}
}

// run sends count produces in batches of depth. Each request is timestamped
// when written and measured when its response is read, so pipelined runs still
// report per-request latency, including time spent waiting in the pipeline.
func (c *client) run(frame []byte, count, depth int) []time.Duration {
	out := make([]time.Duration, 0, count)
	for sent := 0; sent < count; {
		batch := min(depth, count-sent)
		sentAt := make([]time.Time, batch)
		for i := range sentAt {
			sentAt[i] = time.Now()
			if _, err := c.w.Write(frame); err != nil {
				panic(err)
			}
		}
		if err := c.w.Flush(); err != nil {
			panic(err)
		}
		for _, at := range sentAt {
			resp, err := protocol.ReadFrame(c.r)
			if err != nil {
				panic(err)
			}
			if _, err := protocol.DecodeResponse(resp.Payload); err != nil {
				panic(err)
			}
			out = append(out, time.Since(at))
		}
		sent += batch
	}
	return out
}

func main() {
	size, count, mode, n := 1024, 2000, "seq", 64
	if len(os.Args) > 1 {
		size, _ = strconv.Atoi(os.Args[1])
	}
	if len(os.Args) > 2 {
		count, _ = strconv.Atoi(os.Args[2])
	}
	if len(os.Args) > 3 {
		mode = os.Args[3]
	}
	if len(os.Args) > 4 {
		n, _ = strconv.Atoi(os.Args[4])
	}
	if size < 1 {
		size = 1024
	}
	if count < 1 {
		count = 1
	}
	if n < 2 {
		n = 64
	}

	produce := &protocol.ProduceRequest{Topic: topic, Message: make([]byte, size)}
	request := &protocol.Request{APIKey: 1, Version: 1, CorrelationID: 1, Body: produce.Encode()}
	frame := (&protocol.Frame{Payload: request.Encode()}).Encode()

	var lats []time.Duration
	var elapsed time.Duration

	switch mode {
	case "pipe":
		c := dial()
		start := time.Now()
		lats = c.run(frame, count, n)
		elapsed = time.Since(start)
		c.conn.Close()

	case "multi":
		conns := min(n, count)
		clients := make([]*client, conns)
		for i := range clients {
			clients[i] = dial()
		}
		parts := make([][]time.Duration, conns)
		var wg sync.WaitGroup
		start := time.Now()
		for i := range clients {
			per := count / conns
			if i < count%conns {
				per++
			}
			wg.Add(1)
			go func(i, per int) {
				defer wg.Done()
				parts[i] = clients[i].run(frame, per, 1)
			}(i, per)
		}
		wg.Wait()
		elapsed = time.Since(start)
		for _, part := range parts {
			lats = append(lats, part...)
		}
		for _, c := range clients {
			c.conn.Close()
		}

	default:
		mode = "seq"
		c := dial()
		start := time.Now()
		lats = c.run(frame, count, 1)
		elapsed = time.Since(start)
		c.conn.Close()
	}

	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
	p := func(q float64) time.Duration { return lats[int(q*float64(len(lats)-1))] }

	fmt.Printf("%-5s %d msgs x %d B: %.1f MB/s, %.0f msgs/s\n",
		mode, count, size,
		float64(count*size)/1e6/elapsed.Seconds(),
		float64(count)/elapsed.Seconds())
	fmt.Printf("latency: p50=%v p99=%v max=%v\n",
		p(0.50).Round(time.Microsecond),
		p(0.99).Round(time.Microsecond),
		lats[len(lats)-1].Round(time.Microsecond))
}
