// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fakecontainer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	skyrecord "github.com/oursky/skycli/record"
	"github.com/twinj/uuid"
)

// FakeDatabase is a map implementation of Database
type FakeDatabase struct {
	RecordList map[string]map[string]*skyrecord.Record
	AssetList  map[string][]byte
}

func NewFakeDatabase() *FakeDatabase {
	return &FakeDatabase{
		RecordList: make(map[string]map[string]*skyrecord.Record),
		AssetList:  make(map[string][]byte),
	}
}

func fakeDatabaseError() error {
	return fmt.Errorf("FakeDatabase Error.")
}

func (d *FakeDatabase) FetchRecord(recordID string) (*skyrecord.Record, error) {
	args := strings.Split(recordID, "/")
	if len(args) != 2 {
		return nil, fakeDatabaseError()
	}
	recordType := args[0]
	recordKey := args[1]

	record, ok := d.RecordList[recordType][recordKey]
	if !ok {
		return nil, fakeDatabaseError()
	}
	return record, nil
}

func (d *FakeDatabase) QueryRecord(recordType string) ([]*skyrecord.Record, error) {
	var recordList []*skyrecord.Record
	records, ok := d.RecordList[recordType]
	if !ok {
		return nil, nil
	}

	for _, record := range records {
		recordList = append(recordList, record)
	}
	return recordList, nil
}

func (d *FakeDatabase) SaveRecord(r *skyrecord.Record) error {
	// Deep clone the record to prevent changing the original one
	var mod bytes.Buffer
	gob.Register(map[string]interface{}{})
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)
	err := enc.Encode(*r)
	if err != nil {
		return err
	}

	var cpy skyrecord.Record
	err = dec.Decode(&cpy)
	if err != nil {
		return err
	}

	args := strings.Split(cpy.RecordID, "/")
	if len(args) != 2 {
		return fakeDatabaseError()
	}
	recordType := args[0]
	recordKey := args[1]

	if _, ok := d.RecordList[recordType]; !ok {
		d.RecordList[recordType] = make(map[string]*skyrecord.Record)
	}

	// Simulate reserved key added by Skygear
	cpy.Data["_reserved"] = "reserved"

	d.RecordList[recordType][recordKey] = &cpy
	return nil
}

func (d *FakeDatabase) DeleteRecord(recordIDList []string) error {
	for _, recordID := range recordIDList {
		args := strings.Split(recordID, "/")
		if len(args) != 2 {
			return fakeDatabaseError()
		}
		recordType := args[0]
		recordKey := args[1]

		if _, ok := d.RecordList[recordType]; !ok {
			return fakeDatabaseError()
		}

		delete(d.RecordList[recordType], recordKey)
	}
	return nil
}

func (d *FakeDatabase) FetchAsset(assetID string) ([]byte, error) {
	data, ok := d.AssetList[assetID]
	if !ok {
		return nil, fakeDatabaseError()
	}
	return data, nil
}

func (d *FakeDatabase) SaveAsset(path string) (string, error) {
	if path == "err" {
		return "", fakeDatabaseError()
	}

	assetID := uuid.NewV4().String() + path
	bytes := []byte(path)
	d.AssetList[assetID] = bytes

	return assetID, nil
}

func (d *FakeDatabase) CreateColumn(recordType, columnName, columnDef string) error {
	return nil
}

func (d *FakeDatabase) RenameColumn(recordType, oldName, newName string) error {
	return nil
}

func (d *FakeDatabase) DeleteColumn(recordType, columnName string) error {
	return nil
}

func (d *FakeDatabase) FetchSchema() (map[string]interface{}, error) {
	return nil, nil
}
