package container

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	skyrecord "github.com/oursky/skycli/record"
)

type SkyDB interface {
	FetchRecord(string) (*skyrecord.Record, error)
	QueryRecord(string) ([]*skyrecord.Record, error)
	SaveRecord(*skyrecord.Record) error
	DeleteRecord([]string) error
	FetchAsset(string) ([]byte, error)
	SaveAsset(string) (string, error)
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

func (d *Database) FetchAsset(assetID string) (assetData []byte, err error) {
	response, err := d.Container.GetAssetRequest(assetID)
	//fmt.Printf("%+v\n", string(response))
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
