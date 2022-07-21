package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/kbinani/screenshot"
)

var com *exec.Cmd
var stdin io.WriteCloser

func main() {
	takeScreenshot()
	partImgLoader()

	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("read Image")
				takeScreenshot()
				partImgLoader()
			case <-quit:
				ticker.Stop()
				fmt.Println("stopped routine")
				killProcess()
				os.Exit(0)
				return
			}
		}
	}()

	//starts routine reading file
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

func partImgLoader() {
	img, err := loadImage("example.jpg")

	if err != nil {
		log.Fatal("Failed to load image", err)
	}

	colours, err := prominentcolor.Kmeans(img)
	if err != nil {
		fmt.Println("Failed to process image", err)
		return
	}

	var allColors []string

	fmt.Println("Dominant colours:")

	for _, colour := range colours {
		allColors = append(allColors, strconv.FormatUint(uint64(colour.Color.R), 10), strconv.FormatUint(uint64(colour.Color.G), 10), strconv.FormatUint(uint64(colour.Color.B), 10))
	}
	fmt.Println(allColors)

	if len(allColors) != 9 {
		fmt.Println("info: not enough colors found !")
		return
	}

	if com == nil {
		changeDecvicesColor(allColors)
	} else {
		refreshProcessValues(allColors)
	}

}

func convertToJpeg() {

	pngImgFile, err := os.Open("example.png")

	if err != nil {
		fmt.Println("PNG-file.png file not found!")
		os.Exit(1)
	}

	defer pngImgFile.Close()

	// create image from PNG file
	imgSrc, err := png.Decode(pngImgFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create a new Image with the same dimension of PNG image
	newImg := image.NewRGBA(imgSrc.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
	jpgImgFile, err := os.Create("example.jpg")

	if err != nil {
		fmt.Println("Cannot create JPEG-file.jpg !")
		fmt.Println(err)
		os.Exit(1)
	}

	defer jpgImgFile.Close()

	var opt jpeg.Options
	opt.Quality = 80

	err = jpeg.Encode(jpgImgFile, newImg, &opt)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Converted PNG file to JPEG file")

}

func takeScreenshot() {

	n := screenshot.NumActiveDisplays()

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(0)

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}
		fileName := "example.png"
		file, _ := os.Create(fileName)
		defer file.Close()
		png.Encode(file, img)

		fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
	}

	convertToJpeg()
}

func loadImage(fileInput string) (image.Image, error) {
	f, err := os.Open(fileInput)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func changeDecvicesColor(allColors []string) {
	app := "color_pulse_by_device_index.exe"
	arg1 := allColors[0]
	arg2 := allColors[1]
	arg3 := allColors[2]
	arg4 := allColors[3]
	arg5 := allColors[4]
	arg6 := allColors[5]
	arg7 := allColors[6]
	arg8 := allColors[7]
	arg9 := allColors[8]

	com = exec.Command(app, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)

	stdin2, e := com.StdinPipe()
	if e != nil {
		panic(e)
	}
	if err := com.Start(); err != nil {
		log.Fatal(err)
	}

	stdin = stdin2
	fmt.Println("started process successfully")
	return

}

func killProcess() {
	err := stdin.Close()
	if err != nil {
		fmt.Println(err)
	}

	if err := com.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}
	fmt.Println("killed process")
}

//todo pipe to prozess with allColors as input
func refreshProcessValues(allColors []string) {
	theString := allColors[0] + " " + allColors[1] + " " + allColors[2] + " " + allColors[3] + " " + allColors[4] + " " + allColors[5] + " " + allColors[6] + " " + allColors[7] + " " + allColors[8] + "\n"
	fmt.Println("The Colors : ", theString)
	//theString := "255" + " " + "0" + " " + "0" + " " + "0" + " " + "255" + " " + "0" + " " + "0" + " " + "0" + " " + "255" + "\n"
	_, e := stdin.Write([]byte(theString))
	if e != nil {
		fmt.Println("failed stdin exit..")
		killProcess()
		os.Exit(0)
	}
}
