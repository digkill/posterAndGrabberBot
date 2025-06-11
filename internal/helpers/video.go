package helpers

import (
	"bytes"
	"fmt"
	"os/exec"
)

func DownloadVKVideo(videoUrl string) error {
	cmd := exec.Command("yt-dlp", "-P", "./storage/content", videoUrl)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		// Возвращаем ошибку + содержимое stderr для дебага
		return fmt.Errorf("yt-dlp error: %w\nSTDOUT:\n%s\nSTDERR:\n%s",
			err, outBuf.String(), errBuf.String())
	}

	fmt.Println("yt-dlp output:\n", outBuf.String())

	return nil
}
