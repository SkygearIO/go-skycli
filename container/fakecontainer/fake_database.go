package fakecontainer

import (
	"fmt"

	skyrecord "github.com/oursky/skycli/record"
)

type FakeDatabase struct{}

func NewFakeDatabase() *FakeDatabase {
	return &FakeDatabase{}
}

func (d *FakeDatabase) FetchRecord(recordID string) (record *skyrecord.Record, err error) {
	return nil, nil
}

func (d *FakeDatabase) SaveRecord(r *skyrecord.Record) error {
	return nil
}

func (d *FakeDatabase) SaveAsset(path string) (assetID string, err error) {
	if path == "err" {
		return "", fmt.Errorf("Something wrong")
	}

	return "Some ID", nil
}
