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

package container

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	skyrecord "github.com/skygeario/skycli/record"
)

type SkyDB interface {
	FetchRecord(string) (*skyrecord.Record, error)
	QueryRecord(string) ([]*skyrecord.Record, error)
	SaveRecord(*skyrecord.Record) error
	DeleteRecord([]string) error
	FetchAsset(string) ([]byte, error)
	SaveAsset(string) (string, error)

	RenameColumn(string, string, string) error
	DeleteColumn(string, string) error
	CreateColumn(string, string, string) error
	FetchSchema() (map[string]interface{}, error)
}

type Database struct {
	Container  *Container
	DatabaseID string
}

func (d *Database) FetchRecord(recordID string) (record *skyrecord.Record, err error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"ids":         []string{recordID},
	}

	response, err := d.Container.MakeRequest("record:fetch", &request)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok || len(resultArray) < 1 {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	resultData, ok := resultArray[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	if IsError(resultData) {
		serverError := MakeError(resultData)
		err = errors.New(serverError.Message)
		return
	}

	record, err = skyrecord.MakeRecord(resultData)
	return
}

func (d *Database) QueryRecord(recordType string) ([]*skyrecord.Record, error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"record_type": recordType,
	}

	response, err := d.Container.MakeRequest("record:query", &request)
	if err != nil {
		return nil, err
	}

	if response.IsError() {
		requestError := response.Error()
		return nil, errors.New(requestError.Message)
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected server data.")
	}

	var recordList []*skyrecord.Record
	for _, r := range resultArray {
		resultData, ok := r.(map[string]interface{})
		if !ok {
			warn(fmt.Errorf("Unexpected server data."))
			continue
		}

		if IsError(resultData) {
			serverError := MakeError(resultData)
			warn(serverError)
			continue
		}

		record, err := skyrecord.MakeRecord(resultData)
		if err != nil {
			warn(err)
			continue
		}

		recordList = append(recordList, record)
	}

	return recordList, nil
}

func (d *Database) SaveRecord(record *skyrecord.Record) (err error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"records":     []skyrecord.Record{*record},
	}

	response, err := d.Container.MakeRequest("record:save", &request)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok || len(resultArray) < 1 {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	resultData, ok := resultArray[0].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	if IsError(resultData) {
		serverError := MakeError(resultData)
		err = errors.New(serverError.Message)
		return
	}
	return
}

func (d *Database) DeleteRecord(recordIDList []string) error {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"ids":         recordIDList,
	}

	response, err := d.Container.MakeRequest("record:delete", &request)
	if err != nil {
		return err
	}

	if response.IsError() {
		requestError := response.Error()
		return errors.New(requestError.Message)
	}

	resultArray, ok := response.Payload["result"].([]interface{})
	if !ok {
		return fmt.Errorf("Unexpected server data.")
	}

	for i := range resultArray {
		resultData, ok := resultArray[i].(map[string]interface{})
		if !ok {
			warn(fmt.Errorf("Encountered unexpected server data."))
			continue
		}

		if IsError(resultData) {
			serverError := MakeError(resultData)
			warn(serverError)
			continue
		}
	}

	return nil
}

func (d *Database) FetchAsset(assetURL string) (assetData []byte, err error) {
	response, err := d.Container.GetAssetRequest(assetURL)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (d *Database) SaveAsset(path string) (assetID string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return
	}
	filename := info.Name()

	//TODO: Use other library to read mime type from content
	filetype := mime.TypeByExtension(filepath.Ext(path))

	response, err := d.Container.PutAssetRequest(filename, filetype, f)
	if err != nil {
		return
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return
	}

	resultData, ok := response.Payload["result"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	assetID, ok = resultData["$name"].(string)
	if !ok {
		err = fmt.Errorf("Unexpected server data.")
		return
	}

	return
}

func (d *Database) RenameColumn(recordType, oldName, newName string) error {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"record_type": recordType,
		"item_type":   "field",
		"item_name":   oldName,
		"new_name":    newName,
	}

	response, err := d.Container.MakeRequest("schema:rename", &request)
	if err != nil {
		return err
	}

	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return err
	}

	//fmt.Printf("%+v\n", response)
	return nil
}

func (d *Database) DeleteColumn(recordType, columnName string) error {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"record_type": recordType,
		"item_type":   "field",
		"item_name":   columnName,
	}

	response, err := d.Container.MakeRequest("schema:delete", &request)
	if err != nil {
		return err
	}
	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return err
	}

	//fmt.Printf("%+v\n", response)
	return nil
}

func (d *Database) CreateColumn(recordType, columnName, columnDef string) error {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
		"record_types": map[string]interface{}{
			recordType: map[string]interface{}{
				"fields": []map[string]string{
					map[string]string{
						"name": columnName,
						"type": columnDef,
					},
				},
			},
		},
	}

	response, err := d.Container.MakeRequest("schema:create", &request)
	if err != nil {
		return err
	}
	if response.IsError() {
		requestError := response.Error()
		err = errors.New(requestError.Message)
		return err
	}

	//fmt.Printf("%+v\n", response)
	return nil
}

func (d *Database) FetchSchema() (map[string]interface{}, error) {
	request := GenericRequest{}
	request.Payload = map[string]interface{}{
		"database_id": d.DatabaseID,
	}

	response, err := d.Container.MakeRequest("schema:fetch", &request)
	if err != nil {
		return nil, err
	}

	if response.IsError() {
		requestError := response.Error()
		return nil, errors.New(requestError.Message)
	}

	result, ok := response.Payload["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected server data.")
	}

	recordTypes, ok := result["record_types"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected server data.")
	}

	return recordTypes, nil
}
