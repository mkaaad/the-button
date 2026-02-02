package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

/*************** 协议结构 ***************/

type BroadcastMsg struct {
	Type string `json:"type"`
	Data struct {
		UserID    int64 `json:"user_id"`
		Timestamp int64 `json:"timestamp"` // ms
	} `json:"data"`
}

/*************** 全局统计 ***************/

var (
	buttonSent    uint64 // TPS
	broadcastRecv uint64 // QPS

	latID     uint64
	latencies sync.Map // id -> latency(ns)
)

/*************** 主函数 ***************/

func main() {
	wsURL := "ws://127.0.0.1:8080/ws"

	clientCount := 500 // 总连接数
	senderCount := 50  // 既发 button 又接收广播
	senderRate := 20   // 每个 sender 的 TPS
	testTime := 30 * time.Second

	var wg sync.WaitGroup
	start := time.Now()

	// 启动统计打印
	go printRealtimeStats()

	// 启动客户端
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runClient(wsURL, id, id < senderCount, senderRate)
		}(i)
	}

	// 等待结束
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	select {
	case <-time.After(testTime):
	case <-sig:
	}

	fmt.Println("\n========== 测试结束 ==========")
	printFinalStats(start)
}

/*************** 客户端逻辑 ***************/

func runClient(url string, id int, sender bool, rate int) {
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		log.Println("dial error:", err)
		return
	}
	defer conn.Close()

	// sender：并发发送 button
	if sender {
		go startButtonSender(conn, rate)
	}

	// 所有客户端：接收广播
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg BroadcastMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		now := time.Now().UnixNano()
		delay := now - msg.Data.Timestamp*1e6

		id := atomic.AddUint64(&latID, 1)
		latencies.Store(id, delay)

		atomic.AddUint64(&broadcastRecv, 1)
	}
}

/*************** 并发发送 button ***************/

func startButtonSender(conn *websocket.Conn, rate int) {
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(1, []byte("button")); err != nil {
			return
		}
		atomic.AddUint64(&buttonSent, 1)
	}
}

/*************** 实时 TPS / QPS ***************/

func printRealtimeStats() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastBtn, lastRecv uint64

	for range ticker.C {
		btn := atomic.LoadUint64(&buttonSent)
		recv := atomic.LoadUint64(&broadcastRecv)

		fmt.Printf(
			"TPS(button): %d | QPS(broadcast): %d\n",
			btn-lastBtn,
			recv-lastRecv,
		)

		lastBtn = btn
		lastRecv = recv
	}
}

/*************** 最终统计 ***************/

func printFinalStats(start time.Time) {
	var (
		count int
		sum   int64
		max   int64
		list  []int64
	)

	latencies.Range(func(_, v any) bool {
		d := v.(int64)
		list = append(list, d)
		sum += d
		if d > max {
			max = d
		}
		count++
		return true
	})

	if count == 0 {
		fmt.Println("无有效数据")
		return
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i] < list[j]
	})

	fmt.Printf("运行时长: %.1fs\n", time.Since(start).Seconds())
	fmt.Printf("Button 总数: %d\n", atomic.LoadUint64(&buttonSent))
	fmt.Printf("Broadcast 总数: %d\n", atomic.LoadUint64(&broadcastRecv))
	fmt.Printf("平均延迟: %.2f ms\n", float64(sum)/float64(count)/1e6)
	fmt.Printf("P50 延迟: %.2f ms\n", float64(list[count*50/100])/1e6)
	fmt.Printf("P90 延迟: %.2f ms\n", float64(list[count*90/100])/1e6)
	fmt.Printf("P99 延迟: %.2f ms\n", float64(list[count*99/100])/1e6)
	fmt.Printf("最大延迟: %.2f ms\n", float64(max)/1e6)
}
