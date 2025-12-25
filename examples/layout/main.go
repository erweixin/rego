package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/erweixin/rego"
)

// =============================================================================
// Layout Demo - Showcase Rego's powerful layout system
// =============================================================================

// Layout modes
const (
	LayoutClassic   = 0 // Header + Sidebar + Content + Footer
	LayoutSplit     = 1 // Two-column split view
	LayoutGrid      = 2 // Grid-like layout with multiple panels
	LayoutCentered  = 3 // Centered content with modal-like appearance
	LayoutDashboard = 4 // Dashboard with stats cards
)

var layoutNames = []string{
	"Classic",
	"Split View",
	"Grid",
	"Centered",
	"Dashboard",
}

func App(c rego.C) rego.Node {
	layoutMode := rego.Use(c, "layout", 0)

	rego.UseKey(c, func(key rego.Key, r rune) {
		switch key {
		case rego.KeyLeft:
			if layoutMode.Val > 0 {
				layoutMode.Set(layoutMode.Val - 1)
			}
		case rego.KeyRight:
			if layoutMode.Val < len(layoutNames)-1 {
				layoutMode.Set(layoutMode.Val + 1)
			}
		}
		switch r {
		case '1', '2', '3', '4', '5':
			idx := int(r - '1')
			if idx < len(layoutNames) {
				layoutMode.Set(idx)
			}
		case 'q':
			c.Quit()
		}
	})

	return rego.VStack(
		// Top navigation bar
		TopBar(c.Child("topbar"), layoutMode.Val, layoutMode.Set),

		rego.Text(""),

		// Main content based on layout mode
		renderLayout(c, layoutMode.Val),

		rego.Text(""),

		// Bottom status bar
		StatusBar(c.Child("status"), layoutNames[layoutMode.Val]),
	).Padding(1, 2)
}

// =============================================================================
// TopBar - Layout selector
// =============================================================================

func TopBar(c rego.C, activeIndex int, setLayout func(int)) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("📐 Layout Showcase").Bold().Color(rego.Cyan),
			rego.Text("   "),
			rego.For(layoutNames, func(name string, i int) rego.Node {
				text := rego.Text(fmt.Sprintf(" [%d] %s ", i+1, name))
				if i == activeIndex {
					return text.Bold().Color(rego.Black).Background(rego.Cyan)
				}
				return text.Color(rego.Gray)
			}),
			rego.Spacer(),
			rego.Text("[←/→] Switch  [q] Quit").Dim(),
		),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1)
}

// =============================================================================
// StatusBar - Bottom status
// =============================================================================

func StatusBar(c rego.C, layoutName string) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("Current: ").Dim(),
			rego.Text(layoutName).Bold().Color(rego.Cyan),
			rego.Spacer(),
			rego.Text("Rego Layout System Demo").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

// =============================================================================
// Layout Renderer
// =============================================================================

func renderLayout(c rego.C, mode int) rego.Node {
	switch mode {
	case LayoutClassic:
		return ClassicLayout(c.Child("classic"))
	case LayoutSplit:
		return SplitLayout(c.Child("split"))
	case LayoutGrid:
		return GridLayout(c.Child("grid"))
	case LayoutCentered:
		return CenteredLayout(c.Child("centered"))
	case LayoutDashboard:
		return DashboardLayout(c.Child("dashboard"))
	default:
		return rego.Text("Unknown layout")
	}
}

// =============================================================================
// Layout 1: Classic - Header + Sidebar + Content + Footer
// =============================================================================

func ClassicLayout(c rego.C) rego.Node {
	return rego.VStack(
		// Description
		rego.Box(
			rego.Text("Classic Layout: Header + Sidebar (fixed width) + Content (flex) + Footer").Italic().Color(rego.Yellow),
		).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(0, 1),

		rego.Text(""),

		// Main area
		rego.HStack(
			// Sidebar - fixed width
			rego.Box(
				rego.VStack(
					rego.Text("📁 Sidebar").Bold().Color(rego.Cyan),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("• Dashboard"),
					rego.Text("• Projects"),
					rego.Text("• Settings"),
					rego.Text("• Help"),
					rego.Spacer(),
					rego.Text("Width: 24").Dim(),
				),
			).Width(24).Border(rego.BorderSingle).BorderColor(rego.Cyan).Padding(1, 1),

			rego.Text(" "),

			// Content - flexible
			rego.Box(
				rego.VStack(
					rego.Text("📄 Main Content").Bold().Color(rego.Green),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("This content area uses Flex(1) to fill"),
					rego.Text("the remaining horizontal space."),
					rego.Text(""),
					rego.Text("The sidebar has a fixed width of 24,"),
					rego.Text("while this area expands automatically."),
					rego.Spacer(),
					rego.Text("Flex: 1").Dim(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Green).Padding(1, 1).Flex(1),
		).Flex(1),
	)
}

// =============================================================================
// Layout 2: Split View - Two equal columns
// =============================================================================

func SplitLayout(c rego.C) rego.Node {
	return rego.VStack(
		// Description
		rego.Box(
			rego.Text("Split View: Two columns with equal flex (Flex: 1 each)").Italic().Color(rego.Yellow),
		).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(0, 1),

		rego.Text(""),

		// Split panels
		rego.HStack(
			// Left panel
			rego.Box(
				rego.VStack(
					rego.Text("◀ Left Panel").Bold().Color(rego.Blue),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("Source Code"),
					rego.Text(""),
					rego.Box(
						rego.VStack(
							rego.Text("func main() {").Color(rego.Cyan),
							rego.Text("    fmt.Println(\"Hello\")").Color(rego.White),
							rego.Text("}").Color(rego.Cyan),
						),
					).Border(rego.BorderSingle).Padding(0, 1),
					rego.Spacer(),
					rego.Text("Flex: 1").Dim(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Blue).Padding(1, 1).Flex(1),

			rego.Text("  "),

			// Right panel
			rego.Box(
				rego.VStack(
					rego.Text("Right Panel ▶").Bold().Color(rego.Magenta),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("Preview Output"),
					rego.Text(""),
					rego.Box(
						rego.VStack(
							rego.Text("$ go run main.go").Dim(),
							rego.Text("Hello").Color(rego.Green),
						),
					).Border(rego.BorderSingle).Padding(0, 1),
					rego.Spacer(),
					rego.Text("Flex: 1").Dim(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Magenta).Padding(1, 1).Flex(1),
		).Flex(1),
	)
}

// =============================================================================
// Layout 3: Grid - Multiple panels in a grid arrangement
// =============================================================================

func GridLayout(c rego.C) rego.Node {
	return rego.VStack(
		// Description
		rego.Box(
			rego.Text("Grid Layout: Multiple panels using nested HStack/VStack").Italic().Color(rego.Yellow),
		).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(0, 1),

		rego.Text(""),

		// Top row
		rego.HStack(
			GridCell(c.Child("cell1"), "📊 Panel 1", "Top-Left", rego.Cyan),
			rego.Text(" "),
			GridCell(c.Child("cell2"), "📈 Panel 2", "Top-Center", rego.Green),
			rego.Text(" "),
			GridCell(c.Child("cell3"), "📉 Panel 3", "Top-Right", rego.Blue),
		),

		rego.Text(""),

		// Bottom row
		rego.HStack(
			GridCell(c.Child("cell4"), "🔧 Panel 4", "Bottom-Left", rego.Magenta),
			rego.Text(" "),
			// Wide panel spanning conceptually 2 cells
			rego.Box(
				rego.VStack(
					rego.Text("📋 Panel 5 (Wide)").Bold().Color(rego.Yellow),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("This panel uses Flex(2) to take"),
					rego.Text("twice the space of other panels."),
					rego.Spacer(),
					rego.Text("Flex: 2").Dim(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(1, 1).Flex(2),
		),
	).Flex(1)
}

func GridCell(c rego.C, title, position string, color rego.Color) rego.Node {
	return rego.Box(
		rego.VStack(
			rego.Text(title).Bold().Color(color),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),
			rego.Text(position).Dim(),
			rego.Spacer(),
			rego.Text("Flex: 1").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(color).Padding(1, 1).Flex(1)
}

// =============================================================================
// Layout 4: Centered - Modal-like centered content
// =============================================================================

func CenteredLayout(c rego.C) rego.Node {
	return rego.VStack(
		// Description
		rego.Box(
			rego.Text("Centered Layout: Using rego.Center() for modal-like positioning").Italic().Color(rego.Yellow),
		).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(0, 1),

		rego.Text(""),

		// Centered content
		rego.Box(
			rego.Center(
				rego.Box(
					rego.VStack(
						rego.Text("🎯 Centered Modal").Bold().Color(rego.Cyan),
						rego.Divider().Color(rego.Gray),
						rego.Text(""),
						rego.Text("This content is perfectly centered"),
						rego.Text("both horizontally and vertically."),
						rego.Text(""),
						rego.Text("Using rego.Center() component"),
						rego.Text("makes centering effortless!"),
						rego.Text(""),
						rego.HStack(
							rego.Text(" ✓ OK ").Bold().Color(rego.Black).Background(rego.Green),
							rego.Text("  "),
							rego.Text(" ✗ Cancel ").Bold().Color(rego.White).Background(rego.Red),
						),
					),
				).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(2, 4),
			),
		).Flex(1),
	)
}

// =============================================================================
// Layout 5: Dashboard - Stats cards with various arrangements
// =============================================================================

func DashboardLayout(c rego.C) rego.Node {
	return rego.VStack(
		// Description
		rego.Box(
			rego.Text("Dashboard Layout: Stats cards, progress bars, and mixed content").Italic().Color(rego.Yellow),
		).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(0, 1),

		rego.Text(""),

		// Stats row
		rego.HStack(
			StatCard("👥 Users", "1,234", "+12%", rego.Cyan),
			rego.Text(" "),
			StatCard("📦 Orders", "567", "+5%", rego.Green),
			rego.Text(" "),
			StatCard("💰 Revenue", "$12.3K", "+23%", rego.Yellow),
			rego.Text(" "),
			StatCard("📊 Visits", "8,901", "-2%", rego.Magenta),
		),

		rego.Text(""),

		// Charts row
		rego.HStack(
			// Progress section
			rego.Box(
				rego.VStack(
					rego.Text("📈 Progress").Bold().Color(rego.Blue),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					ProgressRow("CPU Usage", 75, rego.Green),
					rego.Text(""),
					ProgressRow("Memory", 45, rego.Cyan),
					rego.Text(""),
					ProgressRow("Disk", 90, rego.Red),
					rego.Text(""),
					ProgressRow("Network", 30, rego.Yellow),
					rego.Spacer(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Blue).Padding(1, 1).Flex(1),

			rego.Text(" "),

			// Activity section
			rego.Box(
				rego.VStack(
					rego.Text("📋 Recent Activity").Bold().Color(rego.Green),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					ActivityRow("User signup", "2 min ago", rego.Cyan),
					ActivityRow("Order placed", "5 min ago", rego.Green),
					ActivityRow("Payment received", "12 min ago", rego.Yellow),
					ActivityRow("Review submitted", "1 hour ago", rego.Magenta),
					ActivityRow("Account updated", "2 hours ago", rego.Gray),
					rego.Spacer(),
				),
			).Border(rego.BorderSingle).BorderColor(rego.Green).Padding(1, 1).Flex(1),
		).Flex(1),
	)
}

func StatCard(title, value, change string, color rego.Color) rego.Node {
	changeColor := rego.Green
	if strings.HasPrefix(change, "-") {
		changeColor = rego.Red
	}

	return rego.Box(
		rego.VStack(
			rego.Text(title).Dim(),
			rego.Text(value).Bold().Color(color),
			rego.Text(change).Color(changeColor),
		),
	).Border(rego.BorderRounded).BorderColor(color).Padding(1, 2).Flex(1)
}

func ProgressRow(label string, percent int, color rego.Color) rego.Node {
	width := 20
	filled := (percent * width) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return rego.HStack(
		rego.Text(fmt.Sprintf("%-10s", label)),
		rego.Text(bar).Color(color),
		rego.Text(fmt.Sprintf(" %3d%%", percent)),
	)
}

func ActivityRow(action, time string, color rego.Color) rego.Node {
	return rego.HStack(
		rego.Text("•").Color(color),
		rego.Text(" " + action),
		rego.Spacer(),
		rego.Text(time).Dim(),
	)
}

func main() {
	if err := rego.Run(App); err != nil {
		log.Fatal(err)
	}
}
