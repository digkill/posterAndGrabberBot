package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func main() {
	// Заменить на свою ссылку на видео
	videoPage := "https://vk.com/video-215383571_456284573"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", videoPage, nil)
	// Если видео приватное или требует авторизации — надо подставить свои cookies!
	// req.Header.Set("Cookie", "remixsid=...; remixlang=0; ...")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	found := false
	doc.Find("video source").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && strings.HasSuffix(src, ".mp4") {
			fmt.Println("MP4 URL:", src)
			found = true
		}
	})
	if !found {
		fmt.Println("Не нашёл прямую mp4 ссылку. Возможно, видео приватное или структура страницы изменилась.")
	}
}
