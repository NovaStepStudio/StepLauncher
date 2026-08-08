package Handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const maxBackgroundWidth = 1920
const maxBackgroundHeight = 1080

func (a *App) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) PickBackgroundFile(kind string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexto no disponible")
	}
	var filters []runtime.FileFilter
	switch kind {
	case "image":
		filters = []runtime.FileFilter{
			{DisplayName: "Imagenes (*.png, *.jpg, *.jpeg, *.webp, *.gif, *.bmp)", Pattern: "*.png;*.jpg;*.jpeg;*.webp;*.gif;*.bmp"},
		}
	case "video":
		filters = []runtime.FileFilter{
			{DisplayName: "Videos (*.mp4, *.gif, *.webm)", Pattern: "*.mp4;*.gif;*.webm"},
		}
	default:
		return "", fmt.Errorf("tipo de fondo no soportado")
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Seleccionar fondo",
		Filters: filters,
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return a.ImportBackground(path, kind)
}

func checkResolution(src, kind string) error {
	ext := strings.ToLower(filepath.Ext(src))
	var w, h int
	switch {
	case kind == "video" && ext == ".gif":
		w, h = gifDimensions(src)
	case kind == "video" && ext == ".mp4":
		w, h = mp4Dimensions(src)
	case kind == "video" && ext == ".webm":
		w, h = webmDimensions(src)
	default:
		f, err := os.Open(src)
		if err != nil {
			return nil
		}
		defer f.Close()
		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return nil
		}
		w, h = cfg.Width, cfg.Height
	}
	if w > 0 && h > 0 && (w > maxBackgroundWidth || h > maxBackgroundHeight) {
		return fmt.Errorf("la resolucion debe ser menor a %dx%d (tu archivo: %dx%d)", maxBackgroundWidth, maxBackgroundHeight, w, h)
	}
	return nil
}

func gifDimensions(src string) (int, int) {
	data, err := os.ReadFile(src)
	if err != nil || len(data) < 10 || !bytes.HasPrefix(data, []byte("GIF")) {
		return 0, 0
	}
	return int(binary.LittleEndian.Uint16(data[6:8])), int(binary.LittleEndian.Uint16(data[8:10]))
}

func mp4Dimensions(src string) (int, int) {
	data, err := os.ReadFile(src)
	if err != nil {
		return 0, 0
	}
	idx := bytes.LastIndex(data, []byte("tkhd"))
	if idx < 4 {
		return 0, 0
	}
	boxStart := idx - 4
	boxSize := int(binary.BigEndian.Uint32(data[boxStart:]))
	content := data[idx+4:]
	if boxSize >= 8 && boxSize <= len(content) {
		content = content[:boxSize-8]
	}
	if len(content) < 12 {
		return 0, 0
	}
	w := int(binary.BigEndian.Uint32(content[len(content)-8:])) >> 16
	h := int(binary.BigEndian.Uint32(content[len(content)-4:])) >> 16
	return w, h
}

func webmDimensions(src string) (int, int) {
	data, err := os.ReadFile(src)
	if err != nil {
		return 0, 0
	}
	w, h := 0, 0
	var walk func(start, end int)
	walk = func(start, end int) {
		p := start
		for p+2 <= end && (w == 0 || h == 0) {
			id, n1 := ebmlVint(data, p)
			if n1 == 0 {
				return
			}
			p += n1
			if p >= end {
				return
			}
			size, n2 := ebmlVint(data, p)
			if n2 == 0 {
				return
			}
			p += n2
			if p+size > end {
				return
			}
			switch id {
			case 0x18538067, 0x1654AE6B, 0xAE, 0x1F43B675:
				walk(p, p+size)
			case 0xB0:
				if p+4 <= end {
					w = int(binary.BigEndian.Uint32(data[p:]))
				}
			case 0xBA:
				if p+4 <= end {
					h = int(binary.BigEndian.Uint32(data[p:]))
				}
			}
			p += size
		}
	}
	walk(0, len(data))
	return w, h
}

func ebmlVint(data []byte, pos int) (int, int) {
	if pos >= len(data) {
		return 0, 0
	}
	b := data[pos]
	mask := byte(0x80)
	length := 1
	for mask > 0 && b&mask == 0 {
		mask >>= 1
		length++
	}
	if mask == 0 {
		return 0, 0
	}
	value := int(b & (mask - 1))
	for i := 1; i < length; i++ {
		if pos+i >= len(data) {
			return 0, 0
		}
		value = value<<8 | int(data[pos+i])
	}
	return value, length
}
