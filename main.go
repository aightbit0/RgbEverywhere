package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/kbinani/screenshot"
	"github.com/lucasb-eyer/go-colorful"
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
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

var rgba_old *image.RGBA

func main() {

	conf := loadJSONConfig("rgbeverywhereconf.json")
	//new Instance
	fmt.Println("started: ", time.Now())
	instance := newInstance(conf)
	instance.startExeProgram()
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
	var img *image.RGBA
	var err error
	img2, err2 := screenshot.Capture(400, 500, 290, 290)
	if err2 != nil {
		fmt.Println(err)
		fmt.Println("failed getting screen")
		return
	}

	if rgba_old != nil {
		coloursCheck, err := prominentcolor.KmeansWithAll(1, img2, 0, 60, prominentcolor.GetDefaultMasks())

		if err != nil {
			fmt.Println("Failed to process image", err)
			return
		}
		cl := colorful.Color{
			R: float64(coloursCheck[0].Color.R),
			G: float64(coloursCheck[0].Color.G),
			B: float64(coloursCheck[0].Color.B),
		}

		coloursCheck2, err := prominentcolor.KmeansWithAll(1, rgba_old, 0, 60, prominentcolor.GetDefaultMasks())
		if err != nil {
			fmt.Println("Failed to process image", err)
			rgba_old = nil
			return
		}

		ctwo := colorful.Color{
			R: float64(coloursCheck2[0].Color.R),
			G: float64(coloursCheck2[0].Color.G),
			B: float64(coloursCheck2[0].Color.B),
		}

		distance := cl.DistanceRgb(ctwo)
		//fmt.Println(cl)
		//fmt.Println(ctwo)
		//fmt.Println(distance)

		if distance > 51 {
			img, err = screenshot.Capture(300, 420, 2260, 600)
			if err != nil {
				fmt.Println(err)
				fmt.Println("failed getting screen")
				return
			}
		}

	} else {
		//fmt.Println("initial")
		img, err = screenshot.Capture(300, 420, 2260, 600)
		if err != nil {
			fmt.Println(err)
			fmt.Println("failed getting screen")
			return
		}
		rgba_old = img2
	}

	if img != nil {
		colours, err := prominentcolor.KmeansWithAll(3, img, 0, 70, prominentcolor.GetDefaultMasks())

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
	rgba_old = img2

}

func (r *MainValues) startExeProgram() {
	app := r.cnf.PathToExe
	c := exec.Command(app)
	c.SysProcAttr = &syscall.SysProcAttr{}
	stdin, err := c.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}

	t := &term{}
	t.cmd = c
	t.stdin = stdin

	err = t.cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	r.command = t
	fmt.Println("started process successfully")
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
