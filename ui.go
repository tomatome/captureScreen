// ui.go
package main

import (
	"bytes"
	"fmt"
	"image"
	"net"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/MikolajMGT/h264decoder/decoder"
)

func StartUI() {
	BitmapCH = make(chan Bitmap, 1000)
	go update()
	show()
}

var (
	dec  *decoder.Decoder
	imag *canvas.Image
)

func show() {
	a := app.New()
	w := a.NewWindow("Hello")
	w.Resize(fyne.NewSize(float32(width), float32(height)))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	imag = canvas.NewImageFromImage(img)
	imag.FillMode = canvas.ImageFillOriginal
	w.SetContent(imag)
	uiClient(nil)
	w.ShowAndRun()
}
func update() {
	go func() {
		dec, _ = decoder.New(decoder.PixelFormatRGB, decoder.H264)
		for {
			select {
			case b := <-BitmapCH:
				paint_bitmap(b)
			}
		}
	}()
}

var num1 float64 = 0
var last1 time.Time

func paint_bitmap(b Bitmap) {
	num1++

	frames, _ := dec.Decode(b.Data)
	if len(frames) == 0 {
		fmt.Println("no frames")
	} else {
		for _, frame := range frames {
			m := frame.ToRGB()
			//m := resize.Resize(uint(b.Width), uint(b.Height), m1, resize.Lanczos3)
			imag.Image = m
			imag.Refresh()
		}
	}

	fmt.Println("fps: ", num1/time.Now().Sub(last1).Seconds())
}

var BitmapCH chan Bitmap

func ui_paint_bitmap(b Bitmap) {
	BitmapCH <- b
}

type Screen struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type Info struct {
	Domain   string `json:"domain"`
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Passwd   string `json:"password"`
	Screen   `json:"screen"`
}

func NewInfo(ip, user, passwd string) (error, *Info) {
	var i Info
	if ip == "" || user == "" || passwd == "" {
		return fmt.Errorf("Must ip/user/passwd"), nil
	}
	t := strings.Split(ip, ":")
	i.Ip = t[0]
	i.Port = "3389"
	if len(t) > 1 {
		i.Port = t[1]
	}
	if strings.Index(user, "\\") != -1 {
		t = strings.Split(user, "\\")
		i.Domain = t[0]
		i.Username = t[len(t)-1]
	} else if strings.Index(user, "/") != -1 {
		t = strings.Split(user, "/")
		i.Domain = t[0]
		i.Username = t[len(t)-1]
	} else {
		i.Username = user
	}

	i.Passwd = passwd

	return nil, &i
}

func GetBitmap(conn net.Conn) {
	last1 = time.Now()
	//n := 0
	for {
		b := Read(conn)
		b1 := Bitmap{0, 0, width, height,
			width, height, 4, false, b}
		ui_paint_bitmap(b1)
	}
}
func uiClient(info *Info) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		err error
	)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", "192.168.18.20", "9999"))
	if err != nil {
		fmt.Println("dail failed, err:", err)
		return err
	}
	b := &bytes.Buffer{}
	WriteUInt32LE(uint32(width), b)
	WriteUInt32LE(uint32(height), b)
	Write(b.Bytes(), conn)
	go GetBitmap(conn)

	return err
}

type Bitmap struct {
	DestLeft     int    `json:"destLeft"`
	DestTop      int    `json:"destTop"`
	DestRight    int    `json:"destRight"`
	DestBottom   int    `json:"destBottom"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	BitsPerPixel int    `json:"bitsPerPixel"`
	IsCompress   bool   `json:"isCompress"`
	Data         []byte `json:"data"`
}
