// d3d.go
package main

import (
	"context"
	"fmt"
	"image"
	"time"
	"unsafe"

	"github.com/gonutz/d3d9"
)

func InitD3D9() (d3d9.DISPLAYMODE, *d3d9.Device) {
	d3d, err := d3d9.Create(d3d9.SDK_VERSION)
	defer d3d.Release()
	if err != nil {
		panic("Failed to bind to d3d9")
	}
	mode, err := d3d.GetAdapterDisplayMode(d3d9.ADAPTER_DEFAULT)
	fmt.Println(mode.Format)

	if mode.Format != d3d9.FMT_X8R8G8B8 && mode.Format != d3d9.FMT_A8R8G8B8 {
		panic("Unknown display mode format")
	}

	// Create device
	device, _, _ := d3d.CreateDevice(
		d3d9.ADAPTER_DEFAULT,
		d3d9.DEVTYPE_HAL,
		0,
		d3d9.CREATE_SOFTWARE_VERTEXPROCESSING,
		d3d9.PRESENT_PARAMETERS{
			Windowed:         1,
			BackBufferCount:  1,
			BackBufferWidth:  mode.Width,
			BackBufferHeight: mode.Height,
			SwapEffect:       d3d9.SWAPEFFECT_COPY,
		},
	)
	//defer device.Release()

	return mode, device
}
func d3d9Screenshot(mode d3d9.DISPLAYMODE, device *d3d9.Device) image.Image {
	surface, _ := device.CreateOffscreenPlainSurface(
		uint(mode.Width),
		uint(mode.Height),
		d3d9.FMT_A8R8G8B8,
		d3d9.POOL_SYSTEMMEM,
		0,
	)
	defer surface.Release()
	err := device.GetFrontBufferData(0, surface)

	check(err)
	r, err := surface.LockRect(nil, 0)
	check(err)
	defer surface.UnlockRect()

	if r.Pitch != int32(mode.Width*4) {
		panic("Weird ass padding bruh")
	}

	// Create image of same size as surface
	img := image.NewRGBA(image.Rect(0, 0, int(mode.Width), int(mode.Height)))
	// Copy the shites
	for i := range img.Pix {
		img.Pix[i] = *((*byte)(unsafe.Pointer(r.PBits + uintptr(i))))
	}
	// Covert ARGB to RGBA
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0], img.Pix[i+2] = img.Pix[i+2], img.Pix[i+0]
	}
	return img
}

func D3dCapture(ctx context.Context, c chan image.Image) {
	mode, device := InitD3D9()
	ticker := time.NewTicker(1000 * time.Millisecond / time.Duration(fps))
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		img := d3d9Screenshot(mode, device)
		c <- img
	}
}
