package commands

import (
	"encoding/json"
	"testing"

	fake "github.com/oursky/skycli/container/fakecontainer"
	skyrecord "github.com/oursky/skycli/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertComplexValue(t *testing.T) {
	promptComplexValue = false

	Convey("Convert Location", t, func() {
		data := map[string]interface{}{"_id": "1234", "loc": "@loc:3.14,2.17"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "geo", "$lat": 3.14, "$lng": 2.17}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)

		var resultData map[string]interface{}
		err = json.Unmarshal([]byte(record.Data["loc"].(string)), &resultData)

		So(err, ShouldBeNil)
		So(resultData, ShouldResemble, expectedData)
	})

	Convey("Convert Reference", t, func() {
		data := map[string]interface{}{"_id": "1234", "ref": "@ref:someref"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "ref", "$id": "someref"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)

		var resultData map[string]interface{}
		err = json.Unmarshal([]byte(record.Data["ref"].(string)), &resultData)

		So(err, ShouldBeNil)
		So(resultData, ShouldResemble, expectedData)
	})

	Convey("Convert String", t, func() {
		data := map[string]interface{}{"_id": "1234", "str": "@str:somestr"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "str", "$str": "somestr"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)

		var resultData map[string]interface{}
		err = json.Unmarshal([]byte(record.Data["str"].(string)), &resultData)

		So(err, ShouldBeNil)
		So(resultData, ShouldResemble, expectedData)
	})

	Convey("Convert two complex value", t, func() {
		data := map[string]interface{}{"_id": "1234", "loc": "@loc:3.14,2.17", "ref": "@ref:someref"}
		record, _ := skyrecord.MakeRecord(data)

		expectedLoc := map[string]interface{}{"$type": "geo", "$lat": 3.14, "$lng": 2.17}
		expectedRef := map[string]interface{}{"$type": "ref", "$id": "someref"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)

		var resultLoc, resultRef map[string]interface{}
		err = json.Unmarshal([]byte(record.Data["loc"].(string)), &resultLoc)
		So(err, ShouldBeNil)
		err = json.Unmarshal([]byte(record.Data["ref"].(string)), &resultRef)
		So(err, ShouldBeNil)

		So(resultLoc, ShouldResemble, expectedLoc)
		So(resultRef, ShouldResemble, expectedRef)
	})
}

func TestUploadAssets(t *testing.T) {
	db := fake.NewFakeDatabase()

	Convey("Skip Asset", t, func() {
		skipAsset = true

		data := map[string]interface{}{"_id": "1234", "file": "@file:somefile"}
		record, _ := skyrecord.MakeRecord(data)

		expectedRecord := &skyrecord.Record{
			RecordID: "1234",
			Data:     map[string]interface{}{},
		}

		err := uploadAssets(db, record, "")
		So(err, ShouldBeNil)
		So(record, ShouldResemble, expectedRecord)
	})

	Convey("Upload success", t, func() {
		skipAsset = false

		data := map[string]interface{}{"_id": "1234", "file": "@file:somefile"}
		record, _ := skyrecord.MakeRecord(data)

		expectedRecord := &skyrecord.Record{
			RecordID: "1234",
			Data:     map[string]interface{}{"file": "@asset:Some ID"},
		}

		err := uploadAssets(db, record, "")
		So(err, ShouldBeNil)
		So(record, ShouldResemble, expectedRecord)
	})

	Convey("Upload failure", t, func() {
		skipAsset = false

		data := map[string]interface{}{"_id": "1234", "file": "@file:err"}
		record, _ := skyrecord.MakeRecord(data)

		err := uploadAssets(db, record, "")
		So(err, ShouldNotBeNil)
	})
}
