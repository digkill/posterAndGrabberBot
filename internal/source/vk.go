package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/digkill/posterAndGrabberBot/internal/helpers"
	"github.com/digkill/posterAndGrabberBot/internal/models"
	"github.com/gookit/goutil/dump"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	groupID    = -215383571 // id паблика, всегда с минусом!
	postsCount = 10         // сколько постов за раз (макс 100)
	apiVersion = "5.199"
)

type VKSource struct {
	URL     string
	VKToken string
}

func (s VKSource) Name() string {
	return "VK"
}

func (s VKSource) Fetch(ctx context.Context) (*models.WallGetResponse, error) {

	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}
	return feed, nil

	/*
		return lo.Map(feed, func(item *models.WallGetResponse, _ int) models.WallGetResponse {
			return models.WallGetResponse{
				Title:      item.Title,
				Link:       item.Link,
				Categories: item.Categories,
				Date:       item.Date,

				Summary: strings.TrimSpace(item.Summary),
			}
		}), nil
	*/

}

func getMaxPhotoUrl(photo *models.Photo) (string, string) {
	maxWidth := 0
	maxUrl := ""
	maxType := ""
	for _, size := range photo.Sizes {
		if size.Width > maxWidth {
			maxWidth = size.Width
			maxUrl = size.Url
			maxType = size.Type
		}
	}
	return maxUrl, maxType
}

func NewVK(vkToken string, url string) VKSource {
	return VKSource{
		VKToken: vkToken,
		URL:     url,
	}
}

func (s VKSource) loadFeed(ctx context.Context, url string) (*models.WallGetResponse, error) {
	var feedCh = make(chan *models.WallGetResponse)
	var errCh = make(chan error)

	go func() {
		feed, err := s.grabbing(url)
		if err != nil {
			errCh <- err
			return
		}
		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case feed := <-feedCh:
		return feed, nil
	}
}

func getMaxVideoThumb(video *models.Video) (string, string) {
	maxWidth := 0
	maxUrl := ""
	maxType := ""
	for _, image := range video.Images { // поле Images — массив превью
		if image.Width > maxWidth {
			maxWidth = image.Width
			maxUrl = image.URL
			maxType = fmt.Sprintf("%dx%d", image.Width, image.Height)
		}
	}
	return maxUrl, maxType
}

func saveVideoLink(url, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer f.Close()
	f.WriteString(url)
	fmt.Println("Сохранил ссылку на видео:", filename)
}

func (s VKSource) grabbing(url string) (*models.WallGetResponse, error) {
	fmt.Println("Starting grabbing...")
	offset := 0

	for {
		url := fmt.Sprintf(
			"https://api.vk.com/method/wall.get?owner_id=%d&count=%d&offset=%d&access_token=%s&v=%s",
			groupID, postsCount, offset, s.VKToken, apiVersion,
		)

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var wallResp models.WallGetResponse
		err = json.Unmarshal(body, &wallResp)
		if err != nil {
			fmt.Println("Error parsing:", string(body))
			panic(err)
		}

		items := wallResp.Response.Items
		if len(items) == 0 {
			fmt.Println("Yet not posts.")
			break
		}
		dump.Println(items)

		for _, post := range items {
			for _, attach := range post.Attachments {
				if attach.Type == "photo" && attach.Photo != nil {
					maxUrl, maxType := getMaxPhotoUrl(attach.Photo)
					if maxUrl != "" {
						filename := fmt.Sprintf("post_%d_photo_%d_%s.jpg", post.ID, attach.Photo.ID, maxType)
						helpers.DownloadPhoto(maxUrl, filename)
						// Чтобы не улететь в бан по скорости
						time.Sleep(200 * time.Millisecond)
					}
				}

				if attach.Type == "video" && attach.Video != nil {
					video := attach.Video
					// Формируем ссылку на просмотр видео ВК
					videoUrl := fmt.Sprintf("https://vk.com/video%d_%d", video.OwnerID, video.ID)

					// Качаем превью (thumbnail) — ищем максимальный размер
					maxUrl, maxType := getMaxVideoThumb(video)
					if maxUrl != "" {
						filename := fmt.Sprintf("post_%d_video_%d_%s.jpg", post.ID, video.ID, maxType)
						helpers.DownloadPhoto(maxUrl, filename)
					}
					// Можно сохранить ссылку на видео во внешний файл:
					saveVideoLink(videoUrl, fmt.Sprintf("post_%d_video_%d.txt", post.ID, video.ID))
				}
			}
		}

		offset += postsCount
		// Ограничим на всякий, если много постов
		if offset > 100 {
			fmt.Println("Enough!")
			break
		}
		fmt.Println("Fetched ", len(items), " items.")
		return &wallResp, nil
	}

	fmt.Println("Finished grabbing...")
	return nil, errors.New("no posts found")
}
