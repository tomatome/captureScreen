// dxgi.go
package main

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/kirides/screencapture/d3d"
)

func DxgiCapture(ctx context.Context, c chan image.Image) {
	device, deviceCtx, err := d3d.NewD3D11Device()
	if err != nil {
		fmt.Printf("Could not create D3D11 Device. %v\n", err)
		return
	}
	defer device.Release()
	defer deviceCtx.Release()
	n := 1
	ddup, err := d3d.NewIDXGIOutputDuplication(device, deviceCtx, uint(n))
	if err != nil {
		fmt.Printf("Err NewIDXGIOutputDuplication: %v\n", err)
		return
	}
	defer ddup.Release()

	ticker := time.NewTicker(1000 * time.Millisecond / time.Duration(fps))
	img := image.NewRGBA(image.Rect(0, 0, RectW, RectH))
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}
		ddup.GetImage(img, 0)
		c <- img
	}
}
