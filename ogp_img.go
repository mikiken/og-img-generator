package main

import (
	"bytes"
	"fmt"
	"html"
	"os"
)

// OGP画像のテンプレートのパス
var ogpImgTemplate = "ogp_img_template.svg"

func GenerateOGPImage(articleTitle string) string {
	// テンプレートファイルを読み込む
	svgContent, err := os.ReadFile(ogpImgTemplate)
	if err != nil {
		fmt.Println(err)
	}
	// 記事タイトルをエスケープする
	escapedTitle := html.EscapeString(articleTitle)
	// 記事タイトルをテンプレートに埋め込む
	svgContent = bytes.Replace(svgContent, []byte("{{.article_title}}"), []byte(escapedTitle), -1)

	return string(svgContent)
}

func main() {
	// テスト用の記事タイトル
	articleTitle := "<h1></h1>を楽に記述するためのツールを作った話"

	// OGP画像を生成
	ogpImage := GenerateOGPImage(articleTitle)

	file, err := os.Create("ogp.svg") // ファイルを作成
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close() // 最後にファイルを閉じる

	file.WriteString(ogpImage) // 文字列を書き込む
}
