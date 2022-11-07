package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/kbinani/screenshot"
)

type Config struct {
	RefreshTime int    `json:"refresh"`
	PathToExe   string `json:"pathToExe"`
	Display     int    `json:"display"`
}

type MainValues struct {
	cnf     *Config
	command *term
}

type term struct {
	cmd *exec.Cmd

	stdout io.ReadCloser
	stderr io.ReadCloser
	stdin  io.WriteCloser
}

var optimzed bool = false

func main() {
	conf := loadJSONConfig("rgbeverywhereconf.json")
	//new Instance
	fmt.Println("started: ", time.Now())
	instance := newInstance(conf)
	go instance.startExeProgram()
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
				if instance.command.cmd != nil {
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
		cnf: conf,
	}

	return &s
}

func (r *MainValues) takeScreenshot() {

	bounds := screenshot.GetDisplayBounds(r.cnf.Display)

	var img *image.RGBA
	var err error

	if optimzed {

		img, err = screenshot.Capture(0, 470, 2560, 500) //screenshot.CaptureRect(bounds)
		if err != nil {
			fmt.Println(err)
			fmt.Println("failed getting screen")
			return
		}

		fileName := fmt.Sprintf("%d_%d.png", bounds.Dx(), bounds.Dy())
		file, _ := os.Create(fileName)
		defer file.Close()
		png.Encode(file, img)

	} else {
		img, err = screenshot.Capture(0, 470, 2560, 500) //screenshot.CaptureRect(bounds)
		if err != nil {
			fmt.Println(err)
			fmt.Println("failed getting screen")
			return
		}
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

	if len(allColors) != 9 {
		fmt.Println("info: not enough colors found !")
		return
	}

	r.refreshProcessValues(allColors)

}

func (r *MainValues) startExeProgram() {
	app := r.cnf.PathToExe
	c := exec.Command(app)
	c.SysProcAttr = &syscall.SysProcAttr{}

	stdin, err := c.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		fmt.Println(err)
	}

	t := &term{}
	t.cmd = c
	t.stderr = stderr
	t.stdout = stdout
	t.stdin = stdin

	err = t.cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	r.command = t
	fmt.Println("started process successfully")

	done := make(chan error)
	go func() { done <- r.command.cmd.Wait() }()
	erro := <-done
	if erro != nil {
		fmt.Println("Non-zero exit code:", erro)
	}
}

func (r *MainValues) refreshProcessValues(allColors []string) {
	theString := allColors[0] + "," + allColors[1] + "," + allColors[2] + "," + allColors[3] + "," + allColors[4] + "," + allColors[5] + "," + allColors[6] + "," + allColors[7] + "," + allColors[8] + "\n"
	_, e := r.command.stdin.Write([]byte(theString))
	if e != nil {
		fmt.Println("failed writing trough stdin pipe")
		fmt.Println(e)
		r.killProcess()
	}
}

func (r *MainValues) killProcess() {
	fmt.Println("killing process")
	err := r.command.stdin.Close()
	if err != nil {
		fmt.Println(err)
	}

	if err := r.command.cmd.Process.Kill(); err != nil {
		fmt.Println("failed to kill process: ", err)
	}
	fmt.Println("ended: ", time.Now())
	os.Exit(0)
}
