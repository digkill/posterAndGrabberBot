package nutsdb

import (
	"fmt"
	"github.com/digkill/posterAndGrabberBot/internal/helpers"
	"github.com/nutsdb/nutsdb"
	"log"
)

type NutsDB struct {
	db *nutsdb.DB
}

// Константы для бакета и структуры данных
const (
	bucketName    = "videos"
	dataStructure = nutsdb.DataStructureList
)

func NewNutsDB() *NutsDB {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("./storage/data/nutsdb"),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &NutsDB{
		db: db,
	}
}

func (n *NutsDB) SaveVideoLink(link string) error {
	return n.db.Update(func(tx *nutsdb.Tx) error {
		// Проверяем и создаём бакет, если нужно
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при создании бакета %s: %v", bucketName, err)
			return err
		}
		log.Printf("Бакет %s готов, добавляем ссылку", bucketName)
		return tx.RPush(bucketName, []byte("pending"), []byte(link))
	})
}

func (n *NutsDB) GetAllPendingVideoLinks() ([]string, error) {
	var links []string
	err := n.db.View(func(tx *nutsdb.Tx) error {
		// Проверяем существование бакета
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при доступе к бакету %s: %v", bucketName, err)
			return err
		}
		items, err := tx.LRange(bucketName, []byte("pending"), 0, -1)
		if err != nil {
			return err
		}
		for _, v := range items {
			links = append(links, string(v))
		}
		return nil
	})
	return links, err
}

func (n *NutsDB) RemoveVideoLink(link string) error {
	return n.db.Update(func(tx *nutsdb.Tx) error {
		// Проверяем существование бакета
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при доступе к бакету %s: %v", bucketName, err)
			return err
		}
		// Удаляет только одну копию (count=1)
		return tx.LRem(bucketName, []byte("pending"), 1, []byte(link))
	})
}

func (n *NutsDB) Close() error {
	return n.db.Close()
}

func (n *NutsDB) InitBuckets() error {
	return n.db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при создании бакета %s: %v", bucketName, err)
			return err
		}
		log.Printf("Бакет %s успешно инициализирован", bucketName)
		return nil
	})
}

func (n *NutsDB) TestCreateAndPush() error {
	return n.db.Update(func(tx *nutsdb.Tx) error {
		// Проверяем существование бакета через NewBucket
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при создании бакета %s: %v", bucketName, err)
			return err
		}
		if err := tx.NewBucket(dataStructure, bucketName); err == nutsdb.ErrBucketAlreadyExist {
			log.Printf("Бакет %s уже существует", bucketName)
		} else {
			log.Printf("Создаю бакет %s", bucketName)
		}

		log.Println("Пробую RPush")
		return tx.RPush(bucketName, []byte("pending"), []byte("test_link"))
	})
}

func (n *NutsDB) IsVideoURLProcessed(url string) bool {
	found := false
	err := n.db.View(func(tx *nutsdb.Tx) error {
		key := helpers.UrlKey(url)
		found, _ = tx.SHasKey(bucketName, key)

		return nil
	})
	if err != nil {
		fmt.Printf("Error IsVideoURLProcesses: %s: %v", bucketName, err)
	}
	return found
}

func (n *NutsDB) MarkVideoURLProcessed(url string) error {

	return n.db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.NewBucket(dataStructure, bucketName); err != nil && err != nutsdb.ErrBucketAlreadyExist {
			log.Printf("Ошибка при создании бакета %s: %v", bucketName, err)
			return err
		}
		log.Printf("Бакет %s успешно инициализирован", bucketName)

		key := helpers.UrlKey(url)
		return tx.Put(bucketName, key, []byte{1}, 0)
	})
}
