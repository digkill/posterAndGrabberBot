package helpers

import (
	"os/exec"
)

func DownloadVKVideo(videoUrl string) error {
	cmd := exec.Command("yt-dlp", "-P", "media", videoUrl)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
