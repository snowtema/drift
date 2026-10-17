package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/snowtema/drift/internal/protocol"
)

func goalModel(text string, done, selected bool) Model {
	focus := FocusList
	if selected {
		focus = FocusDetail
	}
	return Model{
		projects: []protocol.FullProject{{Project: protocol.Project{
			Name: "demo", Status: protocol.StatusActive,
			Goals: []protocol.Goal{{Text: text, Done: done}},
			Notes: []protocol.Note{{Ts: "2026-03-22T12:00:00Z", Text: "Keep notes visible"}},
		}}},
		focus: focus, detailSection: SectionGoals,
		width: 120, height: 30,
	}
}

func TestDetailGoalsFitPanel(t *testing.T) {
	for _, width := range []int{36, 64, 110} {
		for _, selected := range []bool{false, true} {
			for _, done := range []bool{false, true} {
				for name, text := range map[string]string{
					"sentence":  strings.Repeat("Ship the next useful feature ", 12),
					"unbroken":  strings.Repeat("x", 300),
					"generated": strings.Repeat("x", 100*1024),
					"wide":      strings.Repeat("開発目標", 80),
					"unicode":   strings.Repeat("café 👩‍💻 e\u0301 ", 40),
				} {
					t.Run(fmt.Sprintf("%s/width=%d/selected=%t/done=%t", name, width, selected, done), func(t *testing.T) {
						model := goalModel(text, done, selected)
						view := model.renderDetail(width, 20)
						lines := strings.Split(view, "\n")
						if len(lines) != 20 {
							t.Fatalf("detail panel changed height: got %d lines", len(lines))
						}
						row := lines[8]
						if got := lipgloss.Width(row); got > width {
							t.Errorf("goal row exceeds panel: got %d cells, want at most %d", got, width)
						}
						if !strings.Contains(row, "…") {
							t.Errorf("long goal has no truncation marker: %q", row)
						}
						icon := "○"
						if done {
							icon = "✓"
						}
						if !strings.Contains(row, icon) || strings.Contains(row, "›") != selected {
							t.Errorf("goal state marker changed: %q", row)
						}
						if !strings.Contains(view, "Keep notes visible") {
							t.Error("long goal displaced the notes section")
						}
						if model.projects[0].Goals[0].Text != text {
							t.Error("rendering changed the stored goal")
						}
					})
				}
			}
		}
	}
}

func TestLongGoalsPreserveFullViewDimensions(t *testing.T) {
	for _, width := range []int{80, 120, 160} {
		for _, selected := range []bool{false, true} {
			t.Run(fmt.Sprintf("width=%d/selected=%t", width, selected), func(t *testing.T) {
				model := goalModel("Ship it", false, selected)
				model.width = width
				before := model.View()
				if !strings.Contains(before, "Ship it") {
					t.Fatal("short goal is not visible")
				}
				model.projects[0].Goals[0].Text = strings.Repeat("Long goal text ", 100)
				after := model.View()
				if lipgloss.Width(after) != lipgloss.Width(before) || lipgloss.Height(after) != lipgloss.Height(before) {
					t.Errorf("long goal changed layout from %dx%d to %dx%d", lipgloss.Width(before), lipgloss.Height(before), lipgloss.Width(after), lipgloss.Height(after))
				}
			})
		}
	}
}

func BenchmarkLongGoalDetail(b *testing.B) {
	for _, size := range []int{1024, 16 * 1024, 32 * 1024} {
		b.Run(fmt.Sprintf("bytes=%d", size), func(b *testing.B) {
			model := goalModel(strings.Repeat("x", size), false, true)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				model.renderDetail(64, 20)
			}
		})
	}
}
