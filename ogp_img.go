package main

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"regexp"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

// OGP画像のテンプレートのパス
var ogpImgTemplate = "ogp_img_template.svg"
var svg_width = 1200
var svg_height = 630

func getTitleFromMetadata(md_filepath string) string {
	md_content, err := os.ReadFile(md_filepath)
	if err != nil {
		fmt.Println(err)
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

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
	// テンプレートファイルを読み込む
	svgContent, err := os.ReadFile(ogpImgTemplate)
	if err != nil {
		fmt.Println(err)
	}
	// 記事タイトルをエスケープする
	escapedTitle := html.EscapeString(articleTitle)
	// 記事タイトルをテンプレートに埋め込む
	svgContent = bytes.Replace(svgContent, []byte("{{.article_title}}"), []byte(escapedTitle), -1)

	return svgContent
}

func convertSvgToPng(svgContent []byte, svg_width int, svg_height int) []byte {
	// ヘッドレスブラウザを起動
	page, err := rod.New().MustConnect().Page(proto.TargetCreateTarget{})
	if err != nil {
		fmt.Println(err)
	}
	// svgファイルを開く
	if err = page.SetDocumentContent(string(svgContent)); err != nil {
		fmt.Println(err)
	}

	// スクリーンショットを撮る
	img, err := page.MustWaitStable().Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      7.5,
			Y:      7.5,
			Width:  float64(svg_width),
			Height: float64(svg_height),
			Scale:  1,
		},
		FromSurface: true,
	})

	if err != nil {
		fmt.Println(err)
	}

	return img
}

func generateOGPImage(articleTitle string) []byte {
	ogpImage := convertSvgToPng(embedTitleToTemplate(articleTitle), svg_width, svg_height)
	return ogpImage
}

func main() {
	md_filepath := "td4-fpga.md"
	articleTitle := getTitleFromMetadata(md_filepath)

	// OGP画像を生成
	ogpImage := generateOGPImage(articleTitle)

	// OGP画像を保存
	pattern := regexp.MustCompile(`\.md$`)
	file, err := os.Create(pattern.ReplaceAllString(md_filepath, ".png")) // ファイルを作成
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close() // 最後にファイルを閉じる

	file.Write(ogpImage)
}
