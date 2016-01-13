package record

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Record represents data that belongs to an Skygear record
type Record struct {
	RecordID string
	Data     map[string]interface{}
}

// Set sets value to a key in the record
func (r *Record) Set(key string, value interface{}) {
	r.Data[key] = value
}

// Get gets value of a key in the record
func (r *Record) Get(key string) (value interface{}, err error) {
	value, ok := r.Data[key]
	if !ok {
		value = ""
	}
	return
}

// Assign is a convenient method for setting value to a key using
// an experssion syntax.
func (r *Record) Assign(expr string) error {
	pair := strings.SplitN(expr, "=", 2)
	if len(pair) < 2 || pair[0] == "" || pair[1] == "" {
		return fmt.Errorf("Record assign '%s' not in correct format. Expected: key=value", expr)
	}

	if strings.HasPrefix(pair[0], "_") {
		return fmt.Errorf("Cannot set data with reserved key: %s", pair[0])
	}

	r.Set(pair[0], pair[1])
	return nil
}

// CheckRecordID checks if specified Record ID conforms to required format
func CheckRecordID(recordID string) error {
	recordIDParts := strings.SplitN(recordID, "/", 2)
	if len(recordIDParts) < 2 || recordIDParts[0] == "" || recordIDParts[1] == "" {
		return errors.New("Error: Record ID not in correct format.")
	}
	return nil
}

// MakeEmptyRecord creates a record with empty data
func MakeEmptyRecord(recordID string) (record *Record, err error) {
	err = CheckRecordID(recordID)
	if err != nil {
		return
	}

	record = &Record{
		RecordID: recordID,
		Data:     map[string]interface{}{},
	}
	return
}

// MakeRecord creates a record
func MakeRecord(data map[string]interface{}) (record *Record, err error) {
	recordID, ok := data["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("Record data not in expected format: '_id' is not string.")
	}
	// Remove the id from the Data map: it is now stored in RecordID
	delete(data, "_id")

	record = &Record{
		RecordID: recordID,
		Data:     data,
	}

	return record, nil
}

// MarshalJSON marshal a record in JSON representation
func (r *Record) MarshalJSON() ([]byte, error) {
	jsonData := map[string]interface{}{
		"_id": r.RecordID,
	}
	for k, v := range r.Data {
		jsonData[k] = v
	}
	return json.Marshal(jsonData)
}

// UnmarshalJSON unmarshal a record from JSON representation
func (r *Record) UnmarshalJSON(b []byte) error {
	jsonMap := map[string]interface{}{}
	err := json.Unmarshal(b, &jsonMap)
	if err != nil {
		return err
	}

	recordID, ok := jsonMap["_id"].(string)
	if !ok {
		return fmt.Errorf("Record data not in expected format: '_id' is not string.")
	}
	r.RecordID = recordID
	r.Data = jsonMap
	return nil
}

// Validate check whether the record format is valid.
func (r *Record) Validate() error {
	err := CheckRecordID(r.RecordID)
	if err != nil {
		return err
	}

	for idx, _ := range r.Data {
		if strings.HasPrefix(idx, "_") {
			return fmt.Errorf("Cannot set data with reserved key: %s", idx)
		}
	}
	return nil
}
