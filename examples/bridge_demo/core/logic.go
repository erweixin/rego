package core

import (
	"fmt"
	"time"
)

// 定义数据协议（这些不依赖 UI 框架）
type AppState struct {
	Progress int
	Status   string
	Logs     []string
}

type Question struct {
	Title string
	Body  string
}

// Handler 是 Core 定义的外部通信协议（完全解耦）
type Handler interface {
	Update(AppState)
	Ask(Question) bool
}

// Run 是业务主逻辑
func Run(h Handler) {
	state := AppState{Status: "准备中...", Logs: []string{"初始化系统接口..."}}
	h.Update(state)
	time.Sleep(1 * time.Second)

	// 第一阶段：扫描
	state.Status = "正在扫描冗余文件..."
	for i := 0; i <= 100; i += 20 {
		state.Progress = i
		state.Logs = append(state.Logs, fmt.Sprintf("扫描路径 /var/log/sys_%d.log", i))
		h.Update(state)
		time.Sleep(400 * time.Millisecond)
	}

	// 第二阶段：交互请求
	state.Status = "等待用户确认"
	h.Update(state)

	confirmed := h.Ask(Question{
		Title: "清理确认",
		Body:  "发现 2.4GB 可清理空间，是否执行彻底清理？",
	})

	// 第三阶段：根据用户反馈执行
	if confirmed {
		state.Status = "正在清理..."
		state.Logs = append(state.Logs, "用户已确认，开始清理...")
		for i := 0; i <= 100; i += 10 {
			state.Progress = i
			h.Update(state)
			time.Sleep(200 * time.Millisecond)
		}
		state.Status = "清理完成"
		state.Logs = append(state.Logs, "系统优化已完成！")
	} else {
		state.Status = "已取消"
		state.Logs = append(state.Logs, "用户取消了清理操作。")
	}
	h.Update(state)
}

