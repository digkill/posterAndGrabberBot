package nutsdb

import (
	"github.com/nutsdb/nutsdb"
	"log"
)

type NutsDB struct {
	db *nutsdb.DB
}

func NewNutsDB() *NutsDB {

	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("/tmp/nutsdb"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *nutsdb.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	return &NutsDB{
		db: db,
	}
}

func (n *NutsDB) SaveVideoLink(link string) error {
	return n.db.Update(func(tx *nutsdb.Tx) error {
		// "videos" — bucket, "pending" — key
		return tx.RPush("videos", []byte("pending"), []byte(link))
	})
}

func (n *NutsDB) GetAllPendingVideoLinks() ([]string, error) {
	var links []string
	err := n.db.View(func(tx *nutsdb.Tx) error {
		items, err := tx.LRange("videos", []byte("pending"), 0, -1)
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
		// Удаляет только одну копию (count=1)
		return tx.LRem("videos", []byte("pending"), 1, []byte(link))
	})
}
