//go:build windows

package platform

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	cfDIB   = 8
	cfHDROP = 15
)

var (
	user32                         = windows.NewLazySystemDLL("user32.dll")
	shell32                        = windows.NewLazySystemDLL("shell32.dll")
	kernel32                       = windows.NewLazySystemDLL("kernel32.dll")
	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procIsClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	procDragQueryFileW             = shell32.NewProc("DragQueryFileW")
	procGlobalLock                 = kernel32.NewProc("GlobalLock")
	procGlobalUnlock               = kernel32.NewProc("GlobalUnlock")
	procGlobalSize                 = kernel32.NewProc("GlobalSize")
)

func clipboardFiles() ([]string, error) {
	if !openClipboard() {
		return nil, nil
	}
	defer closeClipboard()

	if !isFormatAvailable(cfHDROP) {
		return nil, nil
	}

	handle, _, _ := procGetClipboardData.Call(uintptr(cfHDROP))
	if handle == 0 {
		return nil, nil
	}

	count, _, _ := procDragQueryFileW.Call(handle, 0xFFFFFFFF, 0, 0)
	if count == 0 {
		return nil, nil
	}

	paths := make([]string, 0, count)
	buf := make([]uint16, windows.MAX_PATH)
	for i := uintptr(0); i < count; i++ {
		length, _, _ := procDragQueryFileW.Call(handle, i, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if length == 0 {
			continue
		}
		paths = append(paths, windows.UTF16ToString(buf[:length]))
	}
	return paths, nil
}

func clipboardImageFile() (string, bool, error) {
	if !openClipboard() {
		return "", false, nil
	}
	defer closeClipboard()

	if !isFormatAvailable(cfDIB) {
		return "", false, nil
	}

	handle, _, _ := procGetClipboardData.Call(uintptr(cfDIB))
	if handle == 0 {
		return "", false, nil
	}

	ptr, _, _ := procGlobalLock.Call(handle)
	if ptr == 0 {
		return "", false, fmt.Errorf("GlobalLock failed")
	}
	defer procGlobalUnlock.Call(handle)

	size, _, _ := procGlobalSize.Call(handle)
	if size == 0 {
		return "", false, fmt.Errorf("GlobalSize failed")
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	img, err := dibToImage(data)
	if err != nil {
		return "", false, err
	}

	path := filepath.Join(os.TempDir(), "messenger-clipboard.png")
	f, err := os.Create(path)
	if err != nil {
		return "", false, err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return "", false, err
	}
	if err := f.Close(); err != nil {
		return "", false, err
	}
	return path, true, nil
}

func openClipboard() bool {
	r, _, _ := procOpenClipboard.Call(0)
	return r != 0
}

func closeClipboard() {
	procCloseClipboard.Call()
}

func isFormatAvailable(format int) bool {
	r, _, _ := procIsClipboardFormatAvailable.Call(uintptr(format))
	return r != 0
}

func dibToImage(dib []byte) (image.Image, error) {
	if len(dib) < 40 {
		return nil, fmt.Errorf("invalid dib data")
	}

	width := int32(binary.LittleEndian.Uint32(dib[4:8]))
	height := int32(binary.LittleEndian.Uint32(dib[8:12]))
	bitCount := binary.LittleEndian.Uint16(dib[14:16])

	topDown := height < 0
	if topDown {
		height = -height
	}
	if width <= 0 || height <= 0 || (bitCount != 24 && bitCount != 32) {
		return nil, fmt.Errorf("unsupported dib format")
	}

	headerSize := binary.LittleEndian.Uint32(dib[0:4])
	paletteSize := int(headerSize) - 40
	if paletteSize < 0 {
		paletteSize = 0
	}
	offset := int(headerSize) + paletteSize

	stride := ((int(width)*int(bitCount) + 31) / 32) * 4
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	for y := 0; y < int(height); y++ {
		srcY := y
		if !topDown {
			srcY = int(height) - 1 - y
		}
		rowStart := offset + srcY*stride
		if rowStart+stride > len(dib) {
			return nil, fmt.Errorf("dib row out of range")
		}
		for x := 0; x < int(width); x++ {
			i := rowStart + x*int(bitCount)/8
			if bitCount == 24 {
				b := dib[i]
				g := dib[i+1]
				r := dib[i+2]
				img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			} else {
				b := dib[i]
				g := dib[i+1]
				r := dib[i+2]
				a := dib[i+3]
				img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
			}
		}
	}
	return img, nil
}

// ClipboardTextPath returns a file path if clipboard text points to an existing file.
func ClipboardTextPath(text string) (string, bool) {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "\"")
	if text == "" {
		return "", false
	}
	info, err := os.Stat(text)
	if err != nil || info.IsDir() {
		return "", false
	}
	return text, true
}

var _ = syscall.UTF16PtrFromString
