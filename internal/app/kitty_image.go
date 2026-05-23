package app

import (
	"bytes"
	"fmt"
	"os"
)

// Kitty graphics protocol delete-all command (a=d).
const kittyDeleteAll = "\x1b_Ga=d\x1b\\"

// KittyModeActive reports whether kitty-graphics output is selected and chafa
// (which produces the kitty graphics sequences) is available.
func KittyModeActive() bool {
	return thumbFormat == "kitty" && chafaBinary() != ""
}

// PlaceKittyThumbnail clears prior images and renders jpgPath at the given
// terminal cell coordinates (0-indexed), sized to cols x rows cells. The image
// is produced via chafa's kitty backend and surrounded with cursor save/move/
// restore so it doesn't disturb bubbletea's view position.
func PlaceKittyThumbnail(jpgPath string, col, row, cols, rows int) {
	if !KittyModeActive() || jpgPath == "" || cols <= 0 || rows <= 0 {
		return
	}
	if _, err := os.Stat(jpgPath); err != nil {
		return
	}
	raw := RenderThumbnailANSI(jpgPath, cols, rows)
	if raw == "" {
		return
	}
	var buf bytes.Buffer
	buf.WriteString(kittyDeleteAll)
	buf.WriteString("\x1b[s")
	fmt.Fprintf(&buf, "\x1b[%d;%dH", row+1, col+1)
	buf.WriteString(raw)
	buf.WriteString("\x1b[u")
	os.Stdout.Write(buf.Bytes())
}

// ClearKittyImages removes all kitty graphics from the terminal.
func ClearKittyImages() {
	if chafaBinary() == "" {
		return
	}
	fmt.Fprint(os.Stdout, kittyDeleteAll)
}
