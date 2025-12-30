package rego

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Node 是所有视图节点的接口
type Node interface {
	// render 将节点渲染到屏幕指定区域
	// 返回实际使用的高度
	render(screen tcell.Screen, x, y, width, height int) int
}

// flexNode 是支持 flex 布局的节点接口
type flexNode interface {
	Node
	getFlex() int
	getHeight() int
}

// =============================================================================
// clipScreen - 一个包装 screen 的代理，用于实现裁切和滚动偏移
// =============================================================================

type clipScreen struct {
	tcell.Screen
	viewX, viewY int // 视口起始位置 (终端绝对坐标)
	viewW, viewH int // 视口宽高
	offX, offY   int // 内容偏移量
	runtime      *Runtime
}

func (s *clipScreen) SetContent(x, y int, mainc rune, combc []rune, style tcell.Style) {
	// 计算在视口中的实际坐标
	realX := x + s.offX
	realY := y + s.offY

	// 只有在视口范围内的才真正渲染
	if realX >= s.viewX && realX < s.viewX+s.viewW &&
		realY >= s.viewY && realY < s.viewY+s.viewH {
		s.Screen.SetContent(realX, realY, mainc, combc, style)
	}
}

// 拦截光标显示
func (s *clipScreen) ShowCursor(x, y int) {
	realX := x + s.offX
	realY := y + s.offY
	if realX >= s.viewX && realX < s.viewX+s.viewW &&
		realY >= s.viewY && realY < s.viewY+s.viewH {
		s.Screen.ShowCursor(realX, realY)
	}
}

// =============================================================================
// scrollNode - 滚动容器节点
// =============================================================================

type scrollNode struct {
	ctx            *componentContext
	child          Node
	offY           int
	contentHeight  int
	autoScroll     bool // 是否自动滚动到底部
	flex           int
	scrollTopState *State[int]
}

func (s *scrollNode) render(screen tcell.Screen, x, y, width, height int) int {
	if s.child == nil {
		return 0
	}

	// 1. 重新计算内容总高度（基于当前宽度）
	s.contentHeight = measureNodeHeight(s.child, width-1)

	// 2. 如果开启了自动滚动且内容高度超过视口高度，更新偏移量
	if s.autoScroll && s.contentHeight > height {
		s.offY = s.contentHeight - height
		// 同步回状态，以便手动滚动能从正确位置开始
		if s.scrollTopState != nil {
			s.scrollTopState.Set(s.offY)
		}
	}

	// 3. 渲染内容（带裁切代理）
	proxy := &clipScreen{
		Screen:  screen,
		viewX:   x,
		viewY:   y,
		viewW:   width - 1, // 预留一列给滚动条
		viewH:   height,
		offY:    -s.offY,
		runtime: s.ctx.runtime,
	}
	s.child.render(proxy, x, y, width-1, 1000)

	// 3. 绘制滚动条背景轨道
	scrollbarX := x + width - 1
	for i := 0; i < height; i++ {
		screen.SetContent(scrollbarX, y+i, '│', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}

	// 4. 计算并绘制滚动条滑块 (Thumb)
	if s.contentHeight > height {
		thumbHeight := height * height / s.contentHeight
		if thumbHeight < 1 {
			thumbHeight = 1
		}
		// 计算滑块位置
		thumbPos := s.offY * height / s.contentHeight
		if thumbPos+thumbHeight > height {
			thumbPos = height - thumbHeight
		}

		for i := 0; i < thumbHeight; i++ {
			screen.SetContent(scrollbarX, y+thumbPos+i, '┃', nil, tcell.StyleDefault.Foreground(colorToTcell(Cyan)))
		}
	}

	return height
}

// ScrollBox 创建一个可滚动的容器
func ScrollBox(c C, child Node) *componentNode {
	scrollTop := Use(c, "scrollTop", 0)
	autoScroll := Use(c, "autoScroll", false)
	ctx := c.(*componentContext)

	// 监听鼠标滚轮
	UseMouse(c, func(ev MouseEvent) {
		rect := c.Rect()
		if rect.Contains(ev.X, ev.Y) {
			if ev.Type == MouseEventScrollUp {
				autoScroll.Set(false) // 手动向上滚动时取消自动滚动
				scrollTop.Update(func(v int) int {
					if v > 0 {
						return v - 1
					}
					return 0
				})
			} else if ev.Type == MouseEventScrollDown {
				// 获取子节点的高度
				contentHeight := measureNodeHeight(child, rect.W-1)
				scrollTop.Update(func(v int) int {
					maxScroll := contentHeight - rect.H
					if maxScroll < 0 {
						maxScroll = 0
					}
					if v < maxScroll {
						return v + 1
					} else {
						// 滚到底部了，可以重新开启自动滚动
						autoScroll.Set(true)
						return v
					}
				})
			}
		}
	})

	node := &scrollNode{
		ctx:            ctx,
		child:          child,
		offY:           scrollTop.Val,
		autoScroll:     autoScroll.Val,
		scrollTopState: scrollTop,
	}
	return c.Wrap(node)
}

// TailBox 是一个默认开启自动滚动的 ScrollBox，非常适合日志和聊天界面
func TailBox(c C, child Node) *componentNode {
	// 如果是第一次创建，默认开启自动滚动
	Use(c, "autoScroll", true)
	return ScrollBox(c, child)
}

func (s *scrollNode) AutoScroll(auto bool) *scrollNode {
	s.autoScroll = auto
	// 同步回状态
	autoScroll := Use(s.ctx, "autoScroll", false)
	autoScroll.Set(auto)
	return s
}

func (s *scrollNode) Flex(f int) *scrollNode {
	s.flex = f
	return s
}

func (s *scrollNode) getFlex() int {
	if s.flex > 0 {
		return s.flex
	}
	return 1 // ScrollBox 默认 flex=1
}

func (s *scrollNode) getHeight() int {
	return 0
}

// =============================================================================
// componentNode - 内部节点，用于记录组件在屏幕上的位置
// =============================================================================

type componentNode struct {
	ctx  *componentContext
	node Node
}

func (cn *componentNode) render(screen tcell.Screen, x, y, width, height int) int {
	usedHeight := 0
	if cn.node != nil {
		usedHeight = cn.node.render(screen, x, y, width, height)
	}

	// 记录布局位置（使用实际渲染的高度）
	cn.ctx.rect = Rect{X: x, Y: y, W: width, H: usedHeight}

	return usedHeight
}

func (cn *componentNode) getFlex() int {
	// 如果 inner node 本身有 flex (如 Spacer)，优先使用
	if fn, ok := cn.node.(flexNode); ok && fn.getFlex() > 0 {
		return fn.getFlex()
	}
	// 否则尝试使用 inner node 为 textNode 的样式（兼容旧代码）
	if tn, ok := cn.node.(*textNode); ok {
		return tn.style.flex
	}
	// 最后尝试使用 scrollNode 等设置的 flex
	if sn, ok := cn.node.(interface{ getFlex() int }); ok {
		return sn.getFlex()
	}
	return 0
}

func (cn *componentNode) Flex(f int) *componentNode {
	// 如果 inner node 支持设置 flex，则透传
	type flexSetter interface {
		Flex(int) *scrollNode // 目前只有 scrollNode 有这个方法，返回自己
	}
	if fs, ok := cn.node.(flexSetter); ok {
		fs.Flex(f)
	}
	// 以后可以考虑在 componentNode 本身维护一个 Style 以便通用的 Flex/Padding 支持
	return cn
}

func (cn *componentNode) Padding(top, horizontal int) *componentNode {
	// 暂时只支持透传给 vstackNode 等
	type paddingSetter interface {
		Padding(int, int) *vstackNode
	}
	if ps, ok := cn.node.(paddingSetter); ok {
		ps.Padding(top, horizontal)
	}
	return cn
}

func (cn *componentNode) getHeight() int {
	if fn, ok := cn.node.(flexNode); ok {
		return fn.getHeight()
	}
	if tn, ok := cn.node.(*textNode); ok {
		if tn.style.height > 0 {
			return tn.style.height
		}
	}
	// 如果宽度还不知道（即还没 render 过），默认一个宽度
	w := cn.ctx.rect.W
	if w <= 0 {
		w = 80
	}
	return measureNodeHeight(cn.node, w)
}

// =============================================================================
// Text 节点
// =============================================================================

type textNode struct {
	content string
	style   Style
	wrap    bool
}

// Text 创建一个文本节点
func Text(content string) *textNode {
	return &textNode{
		content: content,
		style:   defaultStyle(),
		wrap:    false,
	}
}

// Wrap 启用或禁用自动换行
func (t *textNode) Wrap(w bool) *textNode {
	t.wrap = w
	return t
}

// Apply 应用样式
func (t *textNode) Apply(s Style) *textNode {
	t.style = s
	return t
}

func (t *textNode) render(screen tcell.Screen, x, y, width, height int) int {
	if height <= 0 {
		return 0
	}

	actualWidth := width
	if t.style.width > 0 && t.style.width < width {
		actualWidth = t.style.width
	}

	style := t.style.toTcell()

	if !t.wrap {
		textWidth := runewidth.StringWidth(t.content)
		startX := x
		switch t.style.align {
		case AlignCenter:
			if textWidth < actualWidth {
				startX = x + (actualWidth-textWidth)/2
			}
		case AlignRight:
			if textWidth < actualWidth {
				startX = x + actualWidth - textWidth
			}
		}

		col := startX
		for _, r := range t.content {
			charWidth := runewidth.RuneWidth(r)
			if col+charWidth > x+actualWidth {
				break
			}
			screen.SetContent(col, y, r, nil, style)
			col += charWidth
		}
		return 1
	}

	// 自动换行逻辑
	currentX := x
	currentY := y
	lines := 1

	for _, r := range t.content {
		charWidth := runewidth.RuneWidth(r)

		// 如果当前行放不下，换行
		if currentX+charWidth > x+width {
			currentX = x
			currentY++
			lines++
			if lines > height {
				break
			}
		}

		// 处理显式的换行符
		if r == '\n' {
			currentX = x
			currentY++
			lines++
			if lines > height {
				break
			}
			continue
		}

		screen.SetContent(currentX, currentY, r, nil, style)
		currentX += charWidth
	}

	return lines
}

// 样式链式方法
func (t *textNode) Bold() *textNode {
	t.style.bold = true
	return t
}

func (t *textNode) Italic() *textNode {
	t.style.italic = true
	return t
}

func (t *textNode) Underline() *textNode {
	t.style.underline = true
	return t
}

func (t *textNode) Dim() *textNode {
	t.style.dim = true
	return t
}

func (t *textNode) Blink() *textNode {
	t.style.blink = true
	return t
}

func (t *textNode) Color(c Color) *textNode {
	t.style.fg = c
	return t
}

func (t *textNode) Background(c Color) *textNode {
	t.style.bg = c
	return t
}

// =============================================================================
// Empty 节点
// =============================================================================

type emptyNode struct{}

// Empty 创建一个空节点
func Empty() *emptyNode {
	return &emptyNode{}
}

func (e *emptyNode) render(screen tcell.Screen, x, y, width, height int) int {
	return 0
}

// =============================================================================
// VStack 节点（支持 Flex）
// =============================================================================

type vstackNode struct {
	children []Node
	style    Style
	gap      int
	justify  Align
}

// VStack 创建一个垂直堆叠布局
func VStack(children ...Node) *vstackNode {
	return &vstackNode{
		children: children,
		style:    defaultStyle(),
		justify:  AlignLeft, // 默认为顶部对齐
	}
}

// Apply 应用样式
func (v *vstackNode) Apply(s Style) *vstackNode {
	v.style = s
	return v
}

// Gap 设置子节点之间的间距
func (v *vstackNode) Gap(g int) *vstackNode {
	v.gap = g
	return v
}

// Justify 设置主轴对齐方式 (AlignLeft=Top, AlignCenter=Center, AlignRight=Bottom)
func (v *vstackNode) Justify(a Align) *vstackNode {
	v.justify = a
	return v
}

// Flex 设置 flex 权重
func (v *vstackNode) Flex(f int) *vstackNode {
	v.style.flex = f
	return v
}

// Height 设置固定高度
func (v *vstackNode) Height(h int) *vstackNode {
	v.style.height = h
	return v
}

// Background 设置背景色
func (v *vstackNode) Background(c Color) *vstackNode {
	v.style.bg = c
	return v
}

// Color 设置前景色 (应用到所有子节点，如果子节点没设的话)
func (v *vstackNode) Color(c Color) *vstackNode {
	v.style.fg = c
	return v
}

func (v *vstackNode) getFlex() int {
	return v.style.flex
}

func (v *vstackNode) getHeight() int {
	if v.style.height > 0 {
		return v.style.height
	}
	// 自动计算需要的高度 (这里没有 width，传入 80 作为默认参考，或者更好的方式是让 getHeight 也接收 width)
	return v.measureHeight(80)
}

// measureHeight 测量 VStack 需要的总高度
func (v *vstackNode) measureHeight(width int) int {
	// 减去水平 padding
	innerWidth := width - (v.style.paddingLeft + v.style.paddingRight)
	if innerWidth <= 0 {
		return v.style.paddingTop + v.style.paddingBottom
	}

	total := 0
	count := 0
	for _, child := range v.children {
		if child == nil {
			continue
		}
		total += measureNodeHeight(child, innerWidth)
		count++
	}
	if count > 1 {
		total += (count - 1) * v.gap
	}
	return total + v.style.paddingTop + v.style.paddingBottom
}

// Padding 设置内边距
func (v *vstackNode) Padding(top, horizontal int) *vstackNode {
	v.style.paddingTop = top
	v.style.paddingBottom = top
	v.style.paddingLeft = horizontal
	v.style.paddingRight = horizontal
	return v
}

func (v *vstackNode) render(screen tcell.Screen, x, y, width, height int) int {
	if len(v.children) == 0 {
		return 0
	}

	// 应用 Padding
	x += v.style.paddingLeft
	y += v.style.paddingTop
	width -= (v.style.paddingLeft + v.style.paddingRight)
	height -= (v.style.paddingTop + v.style.paddingBottom)

	if width <= 0 || height <= 0 {
		return 0
	}

	// 过滤有效子节点
	var children []Node
	for _, child := range v.children {
		if child != nil {
			children = append(children, child)
		}
	}
	if len(children) == 0 {
		return 0
	}

	// 第一遍：计算固定高度和总 flex
	fixedHeight := 0
	totalFlex := 0
	numGaps := len(children) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	fixedHeight += numGaps * v.gap

	for _, child := range children {
		if fn, ok := child.(flexNode); ok && fn.getFlex() > 0 {
			totalFlex += fn.getFlex()
		} else {
			fixedHeight += measureNodeHeight(child, width)
		}
	}

	// 计算每个 flex 单位的高度
	remainingHeight := height - fixedHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}
	flexUnitHeight := 0
	if totalFlex > 0 {
		flexUnitHeight = remainingHeight / totalFlex
	}

	// 计算内容总高度（用于对齐）
	totalContentHeight := fixedHeight
	if totalFlex > 0 {
		totalContentHeight += remainingHeight
	}

	// 计算起始 Y 位置 (Justify)
	currentY := y
	switch v.justify {
	case AlignCenter:
		if totalContentHeight < height {
			currentY = y + (height-totalContentHeight)/2
		}
	case AlignRight: // AlignRight 在垂直方向代表 Bottom
		if totalContentHeight < height {
			currentY = y + (height - totalContentHeight)
		}
	}

	totalUsedHeight := 0
	for i, child := range children {
		if currentY >= y+height {
			break
		}

		// 计算子节点的高度
		childHeight := measureNodeHeight(child, width)
		if fn, ok := child.(flexNode); ok && fn.getFlex() > 0 {
			childHeight = flexUnitHeight * fn.getFlex()
		}

		remainingH := (y + height) - currentY
		if childHeight > remainingH {
			childHeight = remainingH
		}

		usedHeight := child.render(screen, x, currentY, width, childHeight)
		if usedHeight == 0 && childHeight > 0 {
			usedHeight = childHeight
		}
		currentY += usedHeight
		totalUsedHeight += usedHeight

		// 加上间距
		if i < len(children)-1 {
			currentY += v.gap
			totalUsedHeight += v.gap
		}
	}

	return totalUsedHeight
}

// =============================================================================
// HStack 节点（支持 Flex）
// =============================================================================

type hstackNode struct {
	children []Node
	style    Style
	gap      int
	justify  Align
}

// HStack 创建一个水平排列布局
func HStack(children ...Node) *hstackNode {
	return &hstackNode{
		children: children,
		style:    defaultStyle(),
		justify:  AlignLeft, // 默认为左对齐
	}
}

// Apply 应用样式
func (h *hstackNode) Apply(s Style) *hstackNode {
	h.style = s
	return h
}

// Gap 设置子节点之间的间距
func (h *hstackNode) Gap(g int) *hstackNode {
	h.gap = g
	return h
}

// Justify 设置主轴对齐方式
func (h *hstackNode) Justify(a Align) *hstackNode {
	h.justify = a
	return h
}

// Flex 设置 flex 权重
func (h *hstackNode) Flex(f int) *hstackNode {
	h.style.flex = f
	return h
}

// Height 设置固定高度
func (h *hstackNode) Height(ht int) *hstackNode {
	h.style.height = ht
	return h
}

// Padding 设置内边距
func (h *hstackNode) Padding(v, hor int) *hstackNode {
	h.style.paddingTop = v
	h.style.paddingBottom = v
	h.style.paddingLeft = hor
	h.style.paddingRight = hor
	return h
}

// Background 设置背景色
func (h *hstackNode) Background(c Color) *hstackNode {
	h.style.bg = c
	return h
}

// Color 设置前景色
func (h *hstackNode) Color(c Color) *hstackNode {
	h.style.fg = c
	return h
}

func (h *hstackNode) getFlex() int {
	return h.style.flex
}

func (h *hstackNode) getHeight() int {
	return h.style.height
}

func (h *hstackNode) render(screen tcell.Screen, x, y, width, height int) int {
	if len(h.children) == 0 {
		return 0
	}

	// 过滤有效子节点
	var children []Node
	for _, child := range h.children {
		if child != nil {
			children = append(children, child)
		}
	}
	if len(children) == 0 {
		return 0
	}

	// 第一遍：计算固定宽度和总 flex
	fixedWidth := 0
	totalFlex := 0
	numGaps := len(children) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	fixedWidth += numGaps * h.gap

	for _, child := range children {
		childWidth := h.measureWidth(child)
		flex := h.getChildFlex(child)

		if flex > 0 {
			totalFlex += flex
		} else {
			fixedWidth += childWidth
		}
	}

	// 计算每个 flex 单位的宽度
	remainingWidth := width - fixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	flexUnitWidth := 0
	if totalFlex > 0 {
		flexUnitWidth = remainingWidth / totalFlex
	}

	// 计算内容总宽度（用于对齐）
	totalContentWidth := fixedWidth
	if totalFlex > 0 {
		totalContentWidth += remainingWidth
	}

	// 计算起始 X 位置 (Justify)
	currentX := x
	switch h.justify {
	case AlignCenter:
		if totalContentWidth < width {
			currentX = x + (width-totalContentWidth)/2
		}
	case AlignRight:
		if totalContentWidth < width {
			currentX = x + (width - totalContentWidth)
		}
	}

	maxHeight := 0
	for i, child := range children {
		if currentX >= x+width {
			break
		}

		childWidth := h.measureWidth(child)
		flex := h.getChildFlex(child)

		if flex > 0 {
			childWidth = flexUnitWidth * flex
		}

		remainingW := (x + width) - currentX
		if childWidth > remainingW {
			childWidth = remainingW
		}

		// 为子节点计算合适的渲染高度
		childRenderHeight := measureNodeHeight(child, childWidth)
		if childRenderHeight > height {
			childRenderHeight = height
		}
		if childRenderHeight == 0 && height > 0 {
			// 如果是 flex 节点（如 Spacer），使用 HStack 的全高
			if _, ok := child.(flexNode); ok {
				childRenderHeight = height
			}
		}

		usedHeight := child.render(screen, currentX, y, childWidth, childRenderHeight)
		if usedHeight > maxHeight {
			maxHeight = usedHeight
		}
		currentX += childWidth

		// 加上间距
		if i < len(children)-1 {
			currentX += h.gap
		}
	}

	if maxHeight == 0 {
		maxHeight = 1
	}
	return maxHeight
}

// measureWidth 测量节点的宽度
func (h *hstackNode) measureWidth(node Node) int {
	total := 0
	switch n := node.(type) {
	case *textNode:
		if n.style.width > 0 {
			total = n.style.width
		} else {
			total = runewidth.StringWidth(n.content)
		}
	case *boxNode:
		if n.style.width > 0 {
			total = n.style.width
		} else {
			// 递归测量子节点宽度，加上 padding 和 border
			childWidth := 0
			if n.child != nil {
				childWidth = h.measureWidth(n.child)
			}
			borderSize := 0
			if n.style.border != BorderNone {
				borderSize = 1
			}
			total = childWidth + n.style.paddingLeft + n.style.paddingRight + borderSize*2
		}
	case *spacerNode:
		total = 0
	case *cursorNode:
		total = 0 // Cursor 不占用宽度
	case *hstackNode:
		// 嵌套的 HStack
		count := 0
		for _, child := range n.children {
			if child == nil {
				continue
			}
			total += n.measureWidth(child)
			count++
		}
		if count > 1 {
			total += (count - 1) * n.gap
		}
	case *vstackNode:
		// VStack 的宽度取决于最宽的子节点
		maxW := 0
		for _, child := range n.children {
			if child == nil {
				continue
			}
			w := h.measureWidth(child)
			if w > maxW {
				maxW = w
			}
		}
		total = maxW
	case *componentNode:
		total = h.measureWidth(n.node)
	case *whenNode:
		if n.condition && n.node != nil {
			total = h.measureWidth(n.node)
		}
	case *whenElseNode:
		if n.condition && n.trueNode != nil {
			total = h.measureWidth(n.trueNode)
		} else if !n.condition && n.falseNode != nil {
			total = h.measureWidth(n.falseNode)
		}
	case *emptyNode:
		total = 0
	default:
		total = 10
	}
	return total
}

// getChildFlex 获取子节点的 flex 值
func (h *hstackNode) getChildFlex(node Node) int {
	switch n := node.(type) {
	case *textNode:
		return n.style.flex
	case *boxNode:
		return n.style.flex
	case *spacerNode:
		return 1 // Spacer 默认 flex=1
	case *vstackNode:
		return n.style.flex
	case *hstackNode:
		return n.style.flex
	case *componentNode:
		return h.getChildFlex(n.node)
	default:
		return 0
	}
}

// =============================================================================
// When 节点（条件渲染）
// =============================================================================

type whenNode struct {
	condition bool
	node      Node
}

// When 条件渲染：当 condition 为 true 时渲染 node
func When(condition bool, node Node) *whenNode {
	return &whenNode{
		condition: condition,
		node:      node,
	}
}

func (w *whenNode) render(screen tcell.Screen, x, y, width, height int) int {
	if w.condition && w.node != nil {
		return w.node.render(screen, x, y, width, height)
	}
	return 0
}

// =============================================================================
// WhenElse 节点
// =============================================================================

type whenElseNode struct {
	condition bool
	trueNode  Node
	falseNode Node
}

// WhenElse 条件渲染：根据 condition 选择渲染哪个节点
func WhenElse(condition bool, trueNode, falseNode Node) *whenElseNode {
	return &whenElseNode{
		condition: condition,
		trueNode:  trueNode,
		falseNode: falseNode,
	}
}

func (w *whenElseNode) render(screen tcell.Screen, x, y, width, height int) int {
	if w.condition {
		if w.trueNode != nil {
			return w.trueNode.render(screen, x, y, width, height)
		}
	} else {
		if w.falseNode != nil {
			return w.falseNode.render(screen, x, y, width, height)
		}
	}
	return 0
}

// =============================================================================
// For 节点（列表渲染）
// =============================================================================

// For 列表渲染：遍历 items 并用 render 函数渲染每个元素
func For[T any](items []T, render func(item T, index int) Node) Node {
	children := make([]Node, len(items))
	for i, item := range items {
		children[i] = render(item, i)
	}
	return VStack(children...)
}

// =============================================================================
// Spacer 节点（flex=1 的弹性空白）
// =============================================================================

type spacerNode struct{}

// Spacer 创建一个弹性空白节点（默认 flex=1）
func Spacer() *spacerNode {
	return &spacerNode{}
}

func (s *spacerNode) render(screen tcell.Screen, x, y, width, height int) int {
	// Spacer 不渲染任何内容，只占据空间
	return height
}

func (s *spacerNode) getFlex() int {
	return 1
}

func (s *spacerNode) getHeight() int {
	return 0
}

// =============================================================================
// Divider 节点（水平分隔线）
// =============================================================================

type dividerNode struct {
	char  rune
	style Style
}

// Divider 创建一个水平分隔线，自动撑满宽度
func Divider() *dividerNode {
	return &dividerNode{
		char:  '─',
		style: defaultStyle(),
	}
}

// Char 设置分隔线使用的字符
func (d *dividerNode) Char(r rune) *dividerNode {
	d.char = r
	return d
}

// Color 设置分隔线颜色
func (d *dividerNode) Color(c Color) *dividerNode {
	d.style.fg = c
	return d
}

func (d *dividerNode) render(screen tcell.Screen, x, y, width, height int) int {
	if height <= 0 || width <= 0 {
		return 0
	}
	style := d.style.toTcell()
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, d.char, nil, style)
	}
	return 1
}

// =============================================================================
// 布局辅助函数
// =============================================================================

// Center 辅助组件：将内容在可用空间内水平和垂直居中
func Center(child Node) Node {
	return VStack(
		Spacer(),
		HStack(
			Spacer(),
			child,
			Spacer(),
		),
		Spacer(),
	)
}

// =============================================================================
// 辅助函数
// =============================================================================

// StringWidth 计算字符串的显示宽度（考虑中文等宽字符）
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// =============================================================================
// Cursor 节点 - 标记光标位置（用于 IME 输入定位）
// =============================================================================

type cursorNode struct {
	runtime *Runtime
}

// Cursor 创建一个光标标记节点
// 需要传入 C 来访问 runtime
func Cursor(c C) Node {
	ctx := c.(*componentContext)
	return &cursorNode{runtime: ctx.runtime}
}

func (n *cursorNode) render(screen tcell.Screen, x, y, width, height int) int {
	// 标记光标位置
	screen.ShowCursor(x, y)
	return 0 // 不占用空间
}

// measureNodeHeight 测量节点需要的高度
func measureNodeHeight(node Node, width int) int {
	switch n := node.(type) {
	case *textNode:
		if !n.wrap || width <= 0 {
			return 1
		}
		// 计算换行后的高度
		currentX := 0
		lines := 1
		for _, r := range n.content {
			charWidth := runewidth.RuneWidth(r)
			if currentX+charWidth > width {
				currentX = 0
				lines++
			}
			if r == '\n' {
				currentX = 0
				lines++
				continue
			}
			currentX += charWidth
		}
		return lines
	case *vstackNode:
		return n.measureHeight(width)
	case *hstackNode:
		maxH := 0
		// HStack 的子节点宽度分配比较复杂，这里简化处理
		// 实际上 HStack 应该知道每个子节点的宽度分配
		for _, child := range n.children {
			if child == nil {
				continue
			}
			h := measureNodeHeight(child, width/len(n.children)) // 粗略估计
			if h > maxH {
				maxH = h
			}
		}
		return maxH
	case *boxNode:
		if n.style.height > 0 {
			return n.style.height
		}
		// Box 默认返回内容高度 + 边框 + padding
		innerHeight := 1
		if n.child != nil {
			// 减去 padding 和边框占用的宽度
			innerW := width
			if n.style.border != BorderNone {
				innerW -= 2
			}
			innerW -= (n.style.paddingLeft + n.style.paddingRight)
			innerHeight = measureNodeHeight(n.child, innerW)
		}
		// 如果内部内容是弹性高度 (0)，Box 整体也应该是弹性高度 (0)
		if innerHeight == 0 {
			return 0
		}
		borderSize := 0
		if n.style.border != BorderNone {
			borderSize = 2
		}
		return innerHeight + borderSize + n.style.paddingTop + n.style.paddingBottom
	case *whenNode:
		if n.condition && n.node != nil {
			return measureNodeHeight(n.node, width)
		}
		return 0
	case *whenElseNode:
		if n.condition && n.trueNode != nil {
			return measureNodeHeight(n.trueNode, width)
		} else if !n.condition && n.falseNode != nil {
			return measureNodeHeight(n.falseNode, width)
		}
		return 0
	case *spacerNode:
		return 0 // Spacer 不占固定高度，是弹性的
	case *scrollNode:
		return 0 // ScrollBox 应该是弹性的，占据所有可用高度
	case *emptyNode:
		return 0
	case *cursorNode:
		return 0 // Cursor 不占用高度
	case *markdownNode:
		return n.measureHeight(width)
	case *componentNode:
		return measureNodeHeight(n.node, width)
	default:
		return 1
	}
}

// 让 boxNode 实现 flexNode 接口
func (b *boxNode) getFlex() int {
	return b.style.flex
}

func (b *boxNode) getHeight() int {
	return b.style.height
}
