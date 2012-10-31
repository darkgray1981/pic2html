package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"strconv"
	"time"

	_ "code.google.com/p/go.image/bmp"
	_ "code.google.com/p/go.image/tiff"
	"code.google.com/p/graphics-go/graphics"
)

// Struct for holding conversion settings
type parameters struct {
	bgcolor    string
	browser    string
	characters string
	contrast   int
	fontsize   string
	grayscale  int
	texttype   string
	width      int

	intensity string
	textcolor string
	invert    int
}

// Returns default parameters
func defaults() parameters {
	var p parameters

	p.bgcolor = "BLACK"
	p.browser = "firefox"
	p.characters = "01"
	p.contrast = 0
	p.fontsize = "-3"
	p.grayscale = 0
	p.texttype = "sequence"
	p.width = 130

	p.intensity = " `.-:/+osyhdmNM"
	p.invert = 0
	p.textcolor = "BLACK"

	return p
}

// Returns a function that returns the next character wanted
func nextChar(kind, characters string) func() string {
	chars := []rune(characters)
	if kind == "random" {
		rand.Seed(time.Now().UnixNano())
		return func() string {
			i := rand.Int() % len(chars)
			return string(chars[i])
		}
	}
	counter := 0
	return func() string {
		i := counter % len(chars)
		counter++
		return string(chars[i])
	}
}

// Converts image to text version
func convertImage(img image.Image, settings map[string]interface{}) (result []byte, err error) {

	// Set up parameters
	var p parameters
	p = defaults()

	if v, ok := settings["bgcolor"].(string); ok && len(v) > 0 {
		p.bgcolor = v
	}
	if v, ok := settings["browser"].(string); ok && len(v) > 0 {
		p.browser = v
	}
	if v, ok := settings["characters"].(string); ok && len(v) > 0 {
		p.characters = v
	}
	if v, ok := settings["contrast"].(string); ok && len(v) > 0 {
		p.contrast, _ = strconv.Atoi(v)
	}
	if v, ok := settings["fontsize"].(string); ok && len(v) > 0 {
		p.fontsize = v
	}
	if v, ok := settings["grayscale"].(string); ok && len(v) > 0 {
		p.grayscale, _ = strconv.Atoi(v)
	}
	if v, ok := settings["textType"].(string); ok && len(v) > 0 {
		p.texttype = v
	}
	if v, ok := settings["width"].(string); ok && len(v) > 0 {
		p.width, _ = strconv.Atoi(v)
		if p.width < 1 || p.width > 500 {
			p.width = 100
		}
	}

	// Rescale proportions to make output prettier in text
	height := int(float32(p.width) / (float32(img.Bounds().Dx()) / float32(img.Bounds().Dy())))
	if p.browser == "ie" {
		height = height * 65 / 100
	} else {
		height = height * 43 / 100
	}

	// Produce minified image to work with
	start := time.Now()
	var m *image.RGBA
	if p.width < img.Bounds().Dx() && p.width >= 150 && (img.Bounds().Dx()*img.Bounds().Dy()/p.width) < 5000 {
		temp, err := resize(img, p.width, height)
		if err != nil {
			return result, err
		}
		m = temp.(*image.RGBA)
	} else {
		m = image.NewRGBA(image.Rect(0, 0, p.width, height))
		graphics.Scale(m, img)
	}

	// Modify image as required
	if p.grayscale == 1 {
		grayscale(m)
	} else if p.grayscale == 2 {
		monochrome(m)
	}

	// Initialize buffer
	var buffer bytes.Buffer
	buffer.WriteString("<table align=\"center\" cellpadding=\"10\">\n<tr bgcolor=\"" + p.bgcolor + "\"><td>\n\n")
	buffer.WriteString("<!-- IMAGE BEGINS HERE -->\n<font size=\"" + p.fontsize + "\">\n<pre>")

	// Prepare variables
	htmlStart := "<font color=#"
	htmlEnd := "</font>"
	current, previous := "", ""
	next := nextChar(p.texttype, p.characters)
	b := m.Bounds()

	// Loop over all pixels and add HTML-font converted data to the output buffer
	for y := b.Min.Y; y < b.Max.Y; y++ {
		j := m.PixOffset(0, y)
		current = hex(uint(m.Pix[j+0])) + hex(uint(m.Pix[j+1])) + hex(uint(m.Pix[j+2]))
		buffer.WriteString(htmlStart + current + ">" + next())
		previous = current

		for x := b.Min.X + 1; x < b.Max.X; x++ {
			i := m.PixOffset(x, y)
			current = hex(uint(m.Pix[i+0])<<16 | uint(m.Pix[i+1])<<8 | uint(m.Pix[i+2]))
			if previous != current {
				buffer.WriteString(htmlEnd + htmlStart + current + ">" + next())
			} else {
				buffer.WriteString(next())
			}
			previous = current
		}

		previous = ""
		buffer.WriteString(htmlEnd + "<br>")
	}

	// Finish up buffer
	buffer.WriteString("\n</pre></font>")
	buffer.WriteString("\n<!-- IMAGE ENDS HERE -->\n")
	buffer.WriteString("\n<FONT COLOR=LIGHTBLUE SIZE=2>Rendering time: " + time.Since(start).String() + "</FONT><BR>\n")

	result = buffer.Bytes()
	return
}

// Parses image data and wanted parameters from posted submission
func parseImage(req *http.Request) (m image.Image, settings map[string]interface{}, err error) {
	settings = make(map[string]interface{})
	if err := req.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		return m, settings, err
	}
	for key, val := range req.Form {
		settings[key] = val[0]
	}
	if req.MultipartForm != nil && len(req.MultipartForm.File["image"]) != 0 {
		p, err := req.MultipartForm.File["image"][0].Open()
		if err != nil {
			return m, settings, err
		}
		m, _, err = image.Decode(p)
		if err != nil {
			return m, settings, err
		}
	} else {
		return nil, nil, fmt.Errorf("No image submitted.")
	}

	return
}

// Returns error style webpage
func Error(err string) []byte {
	var output string

	output += "<table align=center><tr bgcolor=black><td><font color=lightblue size=4>Error: "
	output += err
	output += "</font></td></tr></table>\n"

	return []byte(output)
}

// Main function that handles pic2html conversion
func pic2html(w http.ResponseWriter, req *http.Request) {

	// Initialize HTML parts
	HTML := struct{ top, bottom, error []byte }{}
	var err1, err2, err3 error
	HTML.top, err1 = ioutil.ReadFile("data/result-top.data")
	HTML.bottom, err2 = ioutil.ReadFile("data/result-bottom.data")
	HTML.error, err3 = ioutil.ReadFile("data/result-error.data")
	if err1 != nil || err2 != nil || err3 != nil {
		log.Fatalln("Error:", err1, err2, err3)
	}

	if req.Method != "POST" {
		w.Write(HTML.top)
		w.Write(Error("No input supplied."))
		w.Write(HTML.error)
		return
	}

	m, settings, err := parseImage(req)
	if err != nil {
		w.Write(HTML.top)
		w.Write(Error(err.Error()))
		w.Write(HTML.error)
		return
	}

	result, err := convertImage(m, settings)
	if err != nil {
		w.Write(HTML.top)
		w.Write(Error(err.Error()))
		w.Write(HTML.error)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(result))
	w.Write(HTML.top)
	w.Write(result)
	w.Write(HTML.bottom)

}

// Launches CGI execution
func main() {
	cgi.Serve(http.HandlerFunc(pic2html))
}
