package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"os"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

// path to template svg
var ogpImgTemplate = "ogp_img_template.svg"

func shouldGenerateOGPImage(md_filepath string) bool {
	md_content, err := os.ReadFile(md_filepath)
	if err != nil {
		fmt.Println(err)
	}

	markdown := goldmark.New(goldmark.WithExtensions(meta.Meta))

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert([]byte(string(md_content)), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	metaData := meta.Get(context)
	shouldGenImg := metaData["autoGenOgpImg"]
	return shouldGenImg.(bool)
}

func getTitleFromMetadata(md_filepath string) string {
	md_content, err := os.ReadFile(md_filepath)
	if err != nil {
		fmt.Println(err)
	}

	markdown := goldmark.New(goldmark.WithExtensions(meta.Meta))

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert([]byte(string(md_content)), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	metaData := meta.Get(context)
	title := metaData["title"]
	return title.(string)
}

func embedTitleToTemplate(articleTitle string) []byte {
	// read svg template
	svgContent, err := os.ReadFile(ogpImgTemplate)
	if err != nil {
		fmt.Println(err)
	}
	// escape article title
	escapedTitle := html.EscapeString(articleTitle)
	// embed article title to svg
	svgContent = bytes.Replace(svgContent, []byte("{{.article_title}}"), []byte(escapedTitle), -1)

	return svgContent
}

func getSvgSize(svgBytes []byte) (width int, height int, err error) {
	type SVG struct {
		XMLName xml.Name `xml:"svg"`
		Width   int      `xml:"width,attr"`
		Height  int      `xml:"height,attr"`
	}

	var svg SVG
	// decode the XML content from the byte slice
	err = xml.Unmarshal(svgBytes, &svg)
	if err != nil {
		return -1, -1, err
	}

	// Return width and height
	return svg.Width, svg.Height, nil
}

func convertToPng(svgContent []byte) []byte {
	// get svg size
	width, height, err := getSvgSize(svgContent)
	if err != nil {
		fmt.Println(err)
	}
	// launch headless browser
	page, err := rod.New().MustConnect().Page(proto.TargetCreateTarget{})
	if err != nil {
		fmt.Println(err)
	}
	// set svg content to page
	if err = page.SetDocumentContent(string(svgContent)); err != nil {
		fmt.Println(err)
	}

	// take screenshot
	img, err := page.MustWaitStable().Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      7.5,
			Y:      7.5,
			Width:  float64(width),
			Height: float64(height),
			Scale:  1,
		},
		FromSurface: true,
	})

	if err != nil {
		fmt.Println(err)
	}

	return img
}

func generatePNG(articleTitle string) []byte {
	return convertToPng(embedTitleToTemplate(articleTitle))
}

func main() {
	// get command line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("No markdown file path is specified.")
		fmt.Println("Usage: ogp_img [markdown file path]")
		os.Exit(1)
	}

	for _, md_filepath := range args {
		// check if OGP image should be generated
		if !shouldGenerateOGPImage(md_filepath) {
			continue
		}

		// get article title from metadata
		articleTitle := getTitleFromMetadata(md_filepath)
		// generate OGP image
		ogpImage := generatePNG(articleTitle)

		// save OGP image
		pngFile, err := os.Create(filepath.Dir(md_filepath) + "/ogp.png")
		if err != nil {
			fmt.Println(err)
		}
		defer pngFile.Close()
		pngFile.Write(ogpImage)
	}
}
