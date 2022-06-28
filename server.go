// server.go
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/tomatome/win"

	"github.com/gen2brain/x264-go"
	"github.com/glycerine/rbuf"
)

func startServer() {
	addr := "0.0.0.0:9999"
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("net.ResovleTCPAddr fail:%s", addr)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {

		log.Println("rpc listening", addr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept error:", err)
			continue
		}

		go handleClient(conn)
	}
}

const fps int = 100
const timespan int = 10

var num float64 = 0
var last time.Time
var RectW, RectH int = 1920, 1080

const (
	HORZRES = 8
	VERTRES = 10
)

func ScreenRect() (image.Rectangle, error) {
	// Get device context of whole screen
	hdc := win.GetDC(0)
	if hdc == 0 {
		return image.Rectangle{}, errors.New("GetDC failed")
	}
	defer win.ReleaseDC(0, hdc)
	//x := win.GetDeviceCaps(hdc, HORZRES)
	//y := win.GetDeviceCaps(hdc, VERTRES)
	x, y := RectW, RectH
	return image.Rect(0, 0, x, y), nil
}

func handleClient(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], false)
			fmt.Printf("panic error: %s \n %s\n", err, string(buf[:n]))
			os.Exit(1)
		}
	}()
	d := Read(conn)
	r := bytes.NewReader(d)
	w, _ := ReadUInt32LE(r)
	h, _ := ReadUInt32LE(r)
	fmt.Println("w:", w, "h:", h)

	rect, _ := ScreenRect()
	buf := rbuf.NewFixedSizeRingBuf(50 * 100 * 100)

	// Initialize h264 encoder
	opts := &x264.Options{
		Width:     rect.Dx(),
		Height:    rect.Dy(),
		FrameRate: fps,
		Tune:      "zerolatency",
		Preset:    "ultrafast",
		Profile:   "baseline",
		LogLevel:  x264.LogInfo,
	}
	enc, _ := x264.NewEncoder(buf, opts)

	defer enc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	imgCaps := make(chan image.Image, 1000)
	//go GdiCapture(ctx, imgCaps)
	//go D3dCapture(ctx, imgCaps)
	go DxgiCapture(ctx, imgCaps)

	last = time.Now()
	for img := range imgCaps {
		enc.Encode(img)
		enc.Flush()
		data := buf.Bytes()
		m, err := Write(data, conn)
		if err != nil {
			fmt.Println("server send error:", err, m)
			cancel()
			return
		}
		buf.Reset()
		num++
		fmt.Println("fps: ", num/time.Now().Sub(last).Seconds())
	}
}
func WriteUInt32LE(data uint32, w io.Writer) (int, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, data)
	return w.Write(b)
}
func ReadUInt32LE(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return 0, nil
	}
	return binary.LittleEndian.Uint32(b), nil
}

func Write(data []byte, w io.Writer) (int, error) {
	WriteUInt32LE(uint32(len(data)), w)
	return w.Write(data)
}

func Read(r io.Reader) []byte {
	n, _ := ReadUInt32LE(r)
	b := make([]byte, n)
	io.ReadFull(r, b)

	return b
}
