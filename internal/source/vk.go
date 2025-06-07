package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SlyMarbo/rss"
	"github.com/digkill/posterAndGrabberBot/internal/helpers"
	"github.com/digkill/posterAndGrabberBot/internal/models"
	"io"
	"net/http"
	"time"
)

const (
	vkToken    = "ТВОЙ_VK_API_TOKEN" // подставь свой токен!
	groupID    = -215383571          // id паблика, всегда с минусом!
	postsCount = 10                  // сколько постов за раз (макс 100)
	apiVersion = "5.199"
)

type VKSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

type WallGetResponse struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			ID          int          `json:"id"`
			Attachments []Attachment `json:"attachments"`
		} `json:"items"`
	} `json:"response"`
}

type Attachment struct {
	Type  string `json:"type"`
	Photo *Photo `json:"photo,omitempty"`
}

type Photo struct {
	ID    int `json:"id"`
	Sizes []struct {
		Type   string `json:"type"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"sizes"`
}

func Fetch() {
	offset := 0
	for {
		url := fmt.Sprintf(
			"https://api.vk.com/method/wall.get?owner_id=%d&count=%d&offset=%d&access_token=%s&v=%s",
			groupID, postsCount, offset, vkToken, apiVersion,
		)

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var wallResp WallGetResponse
		err = json.Unmarshal(body, &wallResp)
		if err != nil {
			fmt.Println("Ошибка парсинга:", string(body))
			panic(err)
		}

		items := wallResp.Response.Items
		if len(items) == 0 {
			fmt.Println("Больше постов нет.")
			break
		}

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
			}
		}

		offset += postsCount
		// Ограничим на всякий, если много постов
		if offset > 1000 {
			fmt.Println("Достаточно для теста! Остановлено.")
			break
		}
	}
}

func getMaxPhotoUrl(photo *Photo) (string, string) {
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

func NewVK(m models.Source) VKSource {
	return VKSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

func (s VKSource) ID() int64 {
	return s.SourceID
}

func (s VKSource) Name() string {
	return s.SourceName
}

func (s VKSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var feedCh = make(chan *rss.Feed)
	var errCh = make(chan error)

	go func() {
		feed, err := rss.Fetch(url)
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
