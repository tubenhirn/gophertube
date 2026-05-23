package app

import (
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var (
	thumbCache     = make(map[string]string)
	thumbCacheLock sync.RWMutex

	chafaPathCache string
	chafaPathOnce  sync.Once

	thumbFormat = "symbols"
)

// SetThumbnailFormat selects the chafa output format. Pass "kitty" to enable
// the Kitty graphics protocol; any other value falls back to "symbols".
func SetThumbnailFormat(kitty bool) {
	if kitty {
		thumbFormat = "kitty"
	} else {
		thumbFormat = "symbols"
	}
}

func chafaBinary() string {
	chafaPathOnce.Do(func() {
		if p, err := exec.LookPath("chafa"); err == nil {
			chafaPathCache = p
		}
	})
	return chafaPathCache
}

// RenderThumbnailANSI returns ANSI symbol output for the image at jpgPath sized
// to fit cols x rows terminal cells. Returns "" if chafa is unavailable, the
// path is missing, or rendering fails. Results are cached per (path, cols, rows).
func RenderThumbnailANSI(jpgPath string, cols, rows int) string {
	if jpgPath == "" || cols <= 0 || rows <= 0 {
		return ""
	}
	bin := chafaBinary()
	if bin == "" {
		return ""
	}
	if _, err := os.Stat(jpgPath); err != nil {
		return ""
	}

	key := jpgPath + "|" + strconv.Itoa(cols) + "x" + strconv.Itoa(rows) + "|" + thumbFormat
	thumbCacheLock.RLock()
	if out, ok := thumbCache[key]; ok {
		thumbCacheLock.RUnlock()
		return out
	}
	thumbCacheLock.RUnlock()

	args := []string{
		"--size", strconv.Itoa(cols) + "x" + strconv.Itoa(rows),
		"--format", thumbFormat,
		"--animate", "off",
	}
	if thumbFormat == "symbols" {
		args = append(args, "--dither", "diffusion")
	}
	args = append(args, jpgPath)
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	s := string(out)

	thumbCacheLock.Lock()
	thumbCache[key] = s
	thumbCacheLock.Unlock()
	return s
}
