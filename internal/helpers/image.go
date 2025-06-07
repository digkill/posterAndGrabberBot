package helpers

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
)

func EncodeImageToBase64(imageBytes []byte, fileMimeType string) (string, error) {

	// Кодируем в base64
	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	// Определяем MIME-тип по расширению
	mimeType := mime.TypeByExtension(fileMimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream" // По умолчанию, если неизвестный тип
	}

	// Формируем data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)

	return dataURL, nil
}

func IsImageOrVideo(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov", ".avi":
		return true
	default:
		return false
	}
}

func DownloadPhoto(url, filename string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка загрузки:", err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Ошибка записи файла:", err)
	} else {
		fmt.Println("Скачано:", filename)
	}
}
