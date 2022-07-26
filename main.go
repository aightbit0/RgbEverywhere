package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/kbinani/screenshot"
)

type Config struct {
	RefreshTime int    `json:"refresh"`
	PathToExe   string `json:"pathToExe"`
	Display     int    `json:"display"`
	Mode        bool   `json:"mode"`
}

type MainValues struct {
	com   *exec.Cmd
	stdin io.WriteCloser
	cnf   *Config
	buff  *bytes.Buffer
}

func main() {
	conf := loadJSONConfig("rgbeverywhereconf.json")
	//new Instance
	instance := newInstance(conf)
	fmt.Println(instance.cnf.RefreshTime)
	instance.takeScreenshot()
	ticker := time.NewTicker(time.Duration(instance.cnf.RefreshTime) * time.Millisecond)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				instance.takeScreenshot()
			case <-quit:
				ticker.Stop()
				fmt.Println("stopped routine")
				if instance.com != nil {
					instance.killProcess()
				}
				os.Exit(0)
				return
			}
		}
	}()

	for {
		fmt.Print("action: -> ")
		scanner1 := bufio.NewScanner(os.Stdin)
		var typ string
		if scanner1.Scan() {
			typ = scanner1.Text()
		}
		if typ == "exit" {
			close(quit)
		}
	}
}

func loadJSONConfig(p string) *Config {
	data, err := os.Open(p)
	if err != nil {
		return nil
	}
	d := json.NewDecoder(data)
	var c Config
	if err := d.Decode(&c); err != nil {
		return nil
	}

	return &c
}

func newInstance(conf *Config) *MainValues {
	s := MainValues{
		com:   nil,
		stdin: nil,
		cnf:   conf,
	}

	return &s
}

func (r *MainValues) takeScreenshot() {

	bounds := screenshot.GetDisplayBounds(r.cnf.Display)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}

	buff := new(bytes.Buffer)

	err = jpeg.Encode(buff, img, nil) //&opt

	if err != nil {
		fmt.Println(err)
		fmt.Println("FAILED Convert PNG file to JPEG file")
		return
	}

	r.partImgLoader(buff)
}

func (r *MainValues) partImgLoader(buffer *bytes.Buffer) {
	img, _, err := image.Decode(buffer)

	if err != nil {
		log.Fatal("Failed to load image", err)
	}

	colours, err := prominentcolor.Kmeans(img)
	if err != nil {
		fmt.Println("Failed to process image", err)
		return
	}

	var allColors []string

	for _, colour := range colours {
		allColors = append(allColors, strconv.FormatUint(uint64(colour.Color.R), 10), strconv.FormatUint(uint64(colour.Color.G), 10), strconv.FormatUint(uint64(colour.Color.B), 10))
	}

	if r.cnf.Mode {
		fmt.Println("Dominant colours:")
		fmt.Println(allColors)
	}

	if len(allColors) != 9 {
		fmt.Println("info: not enough colors found !")
		return
	}

	if r.com != nil {
		r.refreshProcessValues(allColors)
		return
	}
	r.startExeProgram(allColors)

}

func (r *MainValues) startExeProgram(allColors []string) {
	app := r.cnf.PathToExe
	arg1 := allColors[0]
	arg2 := allColors[1]
	arg3 := allColors[2]
	arg4 := allColors[3]
	arg5 := allColors[4]
	arg6 := allColors[5]
	arg7 := allColors[6]
	arg8 := allColors[7]
	arg9 := allColors[8]

	r.com = exec.Command(app, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)

	stdin2, e := r.com.StdinPipe()
	if e != nil {
		panic(e)
	}
	if err := r.com.Start(); err != nil {
		log.Fatal(err)
	}

	r.stdin = stdin2
	fmt.Println("started process successfully")
}

func (r *MainValues) refreshProcessValues(allColors []string) {
	theString := allColors[0] + " " + allColors[1] + " " + allColors[2] + " " + allColors[3] + " " + allColors[4] + " " + allColors[5] + " " + allColors[6] + " " + allColors[7] + " " + allColors[8] + "\n"
	_, e := r.stdin.Write([]byte(theString))
	if e != nil {
		fmt.Println("failed stdin exit..")
		r.killProcess()
		os.Exit(0)
	}
}

func (r *MainValues) killProcess() {
	err := r.stdin.Close()
	if err != nil {
		fmt.Println(err)
	}

	if err := r.com.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}
	fmt.Println("killed process")
}
