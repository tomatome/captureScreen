// gdi.go
package main

import (
	"context"
	"errors"
	"image"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

type windowsScreenshot struct {
	x, y, width, height int
	dc                  win.HDC
	bitmap              win.HBITMAP
}

func (h *windowsScreenshot) CreateImage(rect image.Rectangle) (img *image.RGBA, e error) {
	//img = new(image.RGBA)
	e = errors.New("Cannot create image.RGBA")
	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)
	return img, e
}

//获取句柄窗口图像
func (h *windowsScreenshot) Init(width, height int, hwnd win.HWND) error {
	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		return errors.New("GetDC failed")
	}
	defer win.ReleaseDC(hwnd, hdc)

	dc := win.CreateCompatibleDC(hdc)
	if dc == 0 {
		return errors.New("CreateCompatibleDC failed")
	}

	bitmap := win.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bitmap == 0 {
		win.DeleteDC(dc)
		return errors.New("CreateCompatibleBitmap failed")
	}

	h.dc = dc
	h.bitmap = bitmap
	h.x, h.y, h.width, h.height = 0, 0, width, height

	return nil
}

//获取句柄窗口图像
func (h *windowsScreenshot) Capture(hwnd win.HWND) (*image.RGBA, error) {
	rect := image.Rect(h.x, h.y, h.width, h.height)
	img, err := h.CreateImage(rect)
	if err != nil {
		return nil, err
	}
	hdc := win.GetDC(hwnd)
	if hdc == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer win.ReleaseDC(hwnd, hdc)

	var header win.BITMAPINFOHEADER
	header.BiSize = uint32(unsafe.Sizeof(header))
	header.BiPlanes = 1
	header.BiBitCount = 32
	header.BiWidth = int32(h.width)
	header.BiHeight = int32(-h.height)
	header.BiCompression = win.BI_RGB
	header.BiSizeImage = 0

	// GetDIBits balks at using Go memory on some systems. The MSDN example uses
	// GlobalAlloc, so we'll do that too. See:
	// https://docs.microsoft.com/en-gb/windows/desktop/gdi/capturing-an-image
	bitmapDataSize := uintptr(((int64(h.width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(h.height))
	hmem := win.GlobalAlloc(win.GMEM_MOVEABLE, bitmapDataSize)
	defer win.GlobalFree(hmem)
	memptr := win.GlobalLock(hmem)
	defer win.GlobalUnlock(hmem)

	old := win.SelectObject(h.dc, win.HGDIOBJ(h.bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win.SelectObject(h.dc, old)

	if !win.BitBlt(h.dc, 0, 0, int32(h.width), int32(h.height), hdc, int32(h.x), int32(h.y), win.SRCCOPY|win.CAPTUREBLT) {
		return nil, errors.New("BitBlt failed")
	}

	if win.GetDIBits(hdc, h.bitmap, 0, uint32(h.height), (*uint8)(memptr), (*win.BITMAPINFO)(unsafe.Pointer(&header)), win.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	i := 0
	src := uintptr(memptr)
	for y := 0; y < h.height; y++ {
		for x := 0; x < h.width; x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			// BGRA => RGBA, and set A to 255
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += 4
		}
	}

	return img, nil
}

func GdiCapture(ctx context.Context, c chan image.Image) {
	var h windowsScreenshot
	h.Init(RectW, RectH, 0)

	ticker := time.NewTicker(1000 * time.Millisecond / time.Duration(fps))
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}
		img, _ := h.Capture(0)
		c <- img
	}
}
