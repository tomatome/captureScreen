package main

import "C"
import (
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "net/http/pprof"

	_ "github.com/gen2brain/x264-go/x264c"
)

var width int = 920
var height int = 700

func main() {
	startServer()
	if len(os.Args) == 2 {
		t := strings.Split(os.Args[1], "*")
		width, _ = strconv.Atoi(t[0])
		height, _ = strconv.Atoi(t[1])
	}

	StartUI()
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
