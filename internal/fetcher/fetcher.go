package fetcher

import (
	"context"
	"github.com/digkill/posterAndGrabberBot/internal/models"
	src "github.com/digkill/posterAndGrabberBot/internal/source"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
)

type Source interface {
	Name() string
	Fetch(ctx context.Context) ([]models.WallGetResponse, error)
}

type Fetcher struct {
	vkToken        string
	fetchInterval  time.Duration
	filterKeywords []string
}

func NewFetcher(
	vkToken string,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		vkToken:        vkToken,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {

	var wg sync.WaitGroup

	wg.Add(1)

	go func(source Source) {
		defer wg.Done()

		items, err := source.Fetch(ctx)
		if err != nil {
			log.Printf("[ERROR] failed to fetch items from source %q: %v", source.Name(), err)
			return
		}

		if err := f.processItems(ctx, source, items); err != nil {
			log.Printf("[ERROR] failed to process items from source %q: %v", source.Name(), err)
			return
		}

	}(src.NewVK(f.vkToken))

	wg.Wait()

	return nil
}

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []models.WallGetResponse) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			log.Printf("[INFO] item %q (%s) from source %q should be skipped", item.Title, item.Link, source.Name())
			continue
		}

	}

	return nil
}

func (f *Fetcher) itemShouldBeSkipped(item models.WallGetResponse) bool {
	categoriesSet := hashset.New()

	for _, category := range item.Categories {
		categoriesSet.Add(category)
	}

	for _, keyword := range f.filterKeywords {
		if categoriesSet.Contains(keyword) || strings.Contains(strings.ToLower(item.Title), keyword) {
			return true
		}
	}

	return false
}
