package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// previewPathFn returns the absolute path to the JPG to preview for the given
// choice index, or "" if no preview is available for that choice.
type previewPathFn func(idx int) string

type menuListModel struct {
	title        string
	help         string
	choices      []string
	cursor       int
	selected     string
	back         bool
	exit         bool
	preview      previewPathFn
	width        int
	height       int
	kittyMode    bool
	lastImageKey string
}

func newMenuListModel(title, help string, choices []string) menuListModel {
	return menuListModel{
		title:     title,
		help:      help,
		choices:   choices,
		cursor:    0,
		kittyMode: KittyModeActive(),
	}
}

func (m menuListModel) Init() tea.Cmd {
	return nil
}

func (m menuListModel) layout() (leftW, rightW, contentH, thumbCols, thumbRows int) {
	totalW := m.width
	if totalW <= 0 {
		totalW = 100
	}
	leftW = min(56, totalW/2)
	if leftW < 36 {
		leftW = 36
	}
	rightW = totalW - leftW - 2
	if rightW < 20 {
		rightW = 20
	}
	contentH = m.height - 2
	if contentH < 12 {
		contentH = 12
	}
	thumbCols = leftW - uiPadLeft
	if thumbCols < 24 {
		thumbCols = 24
	}
	thumbRows = 16
	if thumbRows > contentH {
		thumbRows = contentH
	}
	return
}

// syncImage matches the search model behavior: place or clear the kitty image
// to reflect the current cursor selection.
func (m menuListModel) syncImage() (menuListModel, tea.Cmd) {
	if !m.kittyMode || m.preview == nil {
		return m, nil
	}
	wantClear := m.width <= 0 || m.cursor < 0 || m.cursor >= len(m.choices)
	var path string
	var col, row, cols, rows int
	if !wantClear {
		path = m.preview(m.cursor)
		if path == "" {
			wantClear = true
		} else {
			_, _, _, cols, rows = m.layout()
			col = uiPadLeft
			row = uiPadTop
		}
	}
	var key string
	if !wantClear {
		key = fmt.Sprintf("%s|%dx%d@%dx%d", path, cols, rows, col, row)
	}
	if key == m.lastImageKey {
		return m, nil
	}
	m.lastImageKey = key
	if wantClear {
		return m, func() tea.Msg { ClearKittyImages(); return nil }
	}
	return m, func() tea.Msg {
		PlaceKittyThumbnail(path, col, row, cols, rows)
		return nil
	}
}

func (m menuListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.update(msg)
	model, imgCmd := model.syncImage()
	return model, tea.Batch(cmd, imgCmd)
}

func (m menuListModel) update(msg tea.Msg) (menuListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.exit = true
			return m, tea.Quit
		case "esc":
			m.back = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.choices) > 0 {
				m.selected = m.choices[m.cursor]
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuListModel) View() string {
	lines := []string{}
	if m.title != "" {
		lines = append(lines, uiIndent()+textAccent.Render(m.title))
		lines = append(lines, "")
	}
	for i, c := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		cStyle := textPrimary
		if i == m.cursor {
			cStyle = textEmphasis
		}
		lines = append(lines, uiIndent()+textAccent.Render(cursor)+cStyle.Render(c))
	}
	if m.help != "" {
		lines = append(lines, "")
		lines = append(lines, uiIndent()+textMuted.Render(m.help))
	}
	list := strings.Join(lines, "\n")

	if m.preview == nil || m.width <= 0 {
		return WithMargin(list)
	}

	leftW, rightW, contentH, thumbCols, thumbRows := m.layout()

	var previewBlock string
	if m.kittyMode {
		previewBlock = strings.Repeat("\n", thumbRows-1)
	} else {
		path := m.preview(m.cursor)
		if path == "" {
			return WithMargin(list)
		}
		raw := RenderThumbnailANSI(path, thumbCols, thumbRows)
		if raw == "" {
			return WithMargin(list)
		}
		previewBlock = indentLines(strings.TrimRight(raw, "\n"), uiIndent())
	}

	leftBox := lipgloss.NewStyle().Width(leftW).Height(contentH).Align(lipgloss.Left).Render(previewBlock)
	rightBox := lipgloss.NewStyle().Width(rightW).Height(contentH).Align(lipgloss.Left).Render(list)
	return WithMargin(lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox))
}

func runMenuTea(title, help string, choices []string) (selected string, back bool, exit bool, err error) {
	return runMenuTeaWithPreview(title, help, choices, nil)
}

func runMenuTeaWithPreview(title, help string, choices []string, preview previewPathFn) (selected string, back bool, exit bool, err error) {
	defer ClearKittyImages()
	model := newMenuListModel(title, help, choices)
	model.preview = preview
	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", false, false, err
	}
	result := m.(menuListModel)
	return result.selected, result.back, result.exit, nil
}
