package main

import (
	"fmt"

	rego "github.com/erweixin/rego"
)

func App(c rego.C) rego.Node {
	// 全局状态：记录最后一次操作的信息
	lastAction := rego.Use(c, "lastAction", "等待操作...")

	rego.UseKey(c, func(key rego.Key, r rune) {
		if r == 'q' {
			c.Quit()
		}
	})

	// 整个应用的渲染计数
	appRenders := rego.UseRef(c, 0)
	appRenders.Current++

	return c.Wrap(rego.Box(
		rego.VStack(
			rego.Text("🖱️ 局部状态与刷新演示").Bold().Color(rego.Cyan),
			rego.Text(fmt.Sprintf("应用总渲染次数: %d", appRenders.Current)).Dim(),
			rego.Text(""),
			rego.Text(fmt.Sprintf("最后操作: %s", lastAction.Val)).Color(rego.Yellow),
			rego.Text(""),

			rego.HStack(
				// 按钮 1：独立实例
				Button(c.Child("btn1"), "🍎 苹果", func() {
					lastAction.Set("你点击了苹果")
				}),
				rego.Text("  "),
				// 按钮 2：独立实例
				Button(c.Child("btn2"), "🍌 香蕉", func() {
					lastAction.Set("你点击了香蕉")
				}),
			),

			rego.Text(""),

			// 按钮 3：另一个实例
			Button(c.Child("btn3"), "🍇 葡萄", func() {
				lastAction.Set("你点击了葡萄")
			}),

			rego.Text(""),
			HoverZone(c.Child("hoverZone")),

			rego.Spacer(),
			rego.Text("观察：点击某个按钮时，只有该按钮的'点击'计数会增加。").Dim(),
			rego.Text("提示：按 [q] 退出").Dim(),
		),
	).Padding(1, 2).Width(70).Height(26).Border(rego.BorderSingle))
}

func Button(c rego.C, label string, onGlobalClick func()) rego.Node {
	// 组件私有状态：每个按钮实例都有自己的点击计数
	localClicks := rego.Use(c, "clicks", 0)

	// 组件私有引用：记录这个特定组件函数被执行了多少次
	renderCount := rego.UseRef(c, 0)
	renderCount.Current++

	focus := rego.UseFocus(c)

	rego.UseMouse(c, func(ev rego.MouseEvent) {
		if ev.Type == rego.MouseEventClick && ev.Button == rego.MouseButtonLeft {
			if c.Rect().Contains(ev.X, ev.Y) {
				localClicks.Update(func(v int) int { return v + 1 })
				onGlobalClick()
				focus.Focus()
			}
		}
	})

	return c.Wrap(rego.Box(
		rego.VStack(
			rego.Text(label).Bold(),
			rego.Text(fmt.Sprintf("点击:%d", localClicks.Val)).Dim(),
			rego.Text(fmt.Sprintf("渲染:%d", renderCount.Current)).Dim(),
		),
	).
		Width(15).
		Border(rego.BorderSingle).
		BorderColor(If(focus.IsFocused, rego.Cyan, rego.Gray)).
		Background(If(focus.IsFocused, rego.Color(rego.Default), rego.Default)).
		Padding(0, 1))
}

func HoverZone(c rego.C) rego.Node {
	hovered := rego.Use(c, "hovered", false)
	renderCount := rego.UseRef(c, 0)
	renderCount.Current++

	rego.UseMouse(c, func(ev rego.MouseEvent) {
		inRange := c.Rect().Contains(ev.X, ev.Y)
		if inRange != hovered.Val {
			hovered.Set(inRange)
		}
	})

	return c.Wrap(rego.Box(
		rego.HStack(
			rego.Text("探测区域 (Hover)").Color(If(hovered.Val, rego.Green, rego.White)),
			rego.Spacer(),
			rego.Text(fmt.Sprintf("渲染:%d", renderCount.Current)).Dim(),
		),
	).Border(rego.BorderSingle).
		BorderColor(If(hovered.Val, rego.Green, rego.Gray)))
}

func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
