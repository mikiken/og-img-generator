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

func convertToPng(svgContent []byte, width int, height int) []byte {
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
	ogpImage := convertToPng(embedTitleToTemplate(articleTitle), svg_width, svg_height)
	return ogpImage
}

func main() {
	// コマンドライン引数を取得
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("引数に.mdファイルを指定してください")
		os.Exit(1)
	}

	for _, md_filepath := range args {
		// 記事タイトルを取得
		articleTitle := getTitleFromMetadata(md_filepath)

		// OGP画像を生成
		ogpImage := generatePNG(articleTitle)

		// OGP画像を保存
		pattern := regexp.MustCompile(`\.md$`)
		pngFile, err := os.Create(pattern.ReplaceAllString(md_filepath, ".png"))
		if err != nil {
			fmt.Println(err)
		}
		defer pngFile.Close() // 最後にファイルを閉じる

		pngFile.Write(ogpImage)
	}
}
