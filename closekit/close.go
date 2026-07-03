package closekit

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
)

type closer struct {
	id string
	px Priority
	fn func() error
}

var (
	closers []closer
	mutex   sync.Mutex
)

// Add 将资源关闭操作添加到关闭队列中
//
//	按 [P0 - P100] 顺序关闭（相同优先级按添加顺序关闭）
func Add(id string, px Priority, fn func() error) {
	mutex.Lock()
	defer mutex.Unlock()

	closers = append(closers, closer{
		id: id,
		px: px,
		fn: fn,
	})
}

// Close 关闭队列中的资源（重复调用安全，已关闭的资源不会被二次关闭）
func Close() {
	// 加锁快照并清空队列，避免与 Add 并发竞争；
	// 在锁外执行关闭，防止关闭回调内再调用 Add/Close 造成死锁
	mutex.Lock()
	list := closers
	closers = nil
	mutex.Unlock()

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].px < list[j].px
	})

	for _, v := range list {
		fmt.Println("⌛️", "close", v.id, "...")
		if err := v.fn(); err != nil {
			slog.Error("close "+v.id+" failed", "error", err)
		}
	}
}
