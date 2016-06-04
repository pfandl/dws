package dws

import (
//	"github.com/vbatts/go-btrfs"
)

type BackingStoreInterface interface {
	Init(path string) error
	Create(path string) error
	Backup(path string) error
	Snapshot(path string) error
}

type BtrfsBackingStore struct {
	BackingStoreInterface
}

func (b *BtrfsBackingStore) Init() error {

	return nil
}
