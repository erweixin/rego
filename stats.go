package rego

import (
	"fmt"
	"time"
)

// Stats 返回一个显示 FPS 和性能信息的组件。
// 它会自动启动一个后台计时器以确保在界面静止时也能更新 FPS 数值。
func Stats(c C) Node {
	fps := Use(c, "fps", 0.0)
	lastTime := UseRef(c, time.Now())
	frames := UseRef(c, 0)

	// 核心修复 1：增加一个定时刷新效应。
	// 确保即使在没有用户交互时，FPS 计数器也能每秒至少更新一次。
	UseEffect(c, func() func() {
		stop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					c.Refresh()
				case <-stop:
					return
				}
			}
		}()
		return func() {
			close(stop)
		}
	})

	// 核心修复 2：更加精确的 FPS 计算。
	// 在每次 render 被调用时累加帧数。
	(*frames).Current++

	now := time.Now()
	diff := now.Sub((*lastTime).Current)

	// 每秒结算一次
	if diff >= time.Second {
		// 计算这一秒内实际发生的渲染次数
		actualFps := float64((*frames).Current) / diff.Seconds()
		fps.Set(actualFps)

		// 重置计数器
		(*frames).Current = 0
		(*lastTime).Current = now
	}

	return Text(fmt.Sprintf(" FPS: %.1f ", fps.Val)).
		Background(Blue).
		Color(White).
		Bold()
}
