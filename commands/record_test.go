package commands

import (
	"reflect"
	"regexp"
	"testing"

	fake "github.com/oursky/skycli/container/fakecontainer"
	skyrecord "github.com/oursky/skycli/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertComplexValue(t *testing.T) {
	forceConvertComplexValue = true

	Convey("Convert Location", t, func() {
		data := map[string]interface{}{"_id": "1234", "loc": "@loc:3.14,2.17"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "geo", "$lat": 3.14, "$lng": 2.17}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)
		So(record.Data["loc"], ShouldResemble, expectedData)
	})

	Convey("Convert Reference", t, func() {
		data := map[string]interface{}{"_id": "1234", "ref": "@ref:someref"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "ref", "$id": "someref"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)
		So(record.Data["ref"], ShouldResemble, expectedData)
	})

	Convey("Convert String", t, func() {
		data := map[string]interface{}{"_id": "1234", "str": "@str:somestr"}
		record, _ := skyrecord.MakeRecord(data)

		expectedData := map[string]interface{}{"$type": "str", "$str": "somestr"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)
		So(record.Data["str"], ShouldResemble, expectedData)
	})

	Convey("Convert two complex value", t, func() {
		data := map[string]interface{}{"_id": "1234", "loc": "@loc:3.14,2.17", "ref": "@ref:someref"}
		record, _ := skyrecord.MakeRecord(data)

		expectedLoc := map[string]interface{}{"$type": "geo", "$lat": 3.14, "$lng": 2.17}
		expectedRef := map[string]interface{}{"$type": "ref", "$id": "someref"}

		err := convertComplexValue(record)
		So(err, ShouldBeNil)
		So(record.Data["loc"], ShouldResemble, expectedLoc)
		So(record.Data["ref"], ShouldResemble, expectedRef)
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

		err := uploadAssets(db, record, "")
		So(err, ShouldBeNil)

		fileMap, ok := record.Data["file"].(map[string]interface{})
		So(ok, ShouldBeTrue)

		fileType, ok := fileMap["$type"].(string)
		So(ok, ShouldBeTrue)
		So(fileType, ShouldEqual, "asset")

		fileName, ok := fileMap["$name"].(string)
		So(ok, ShouldBeTrue)

		ok = regexp.MustCompile(".*somefile$").MatchString(fileName)
		So(ok, ShouldBeTrue)
	})

	Convey("Upload failure", t, func() {
		skipAsset = false

		data := map[string]interface{}{"_id": "1234", "file": "@file:err"}
		record, _ := skyrecord.MakeRecord(data)

		err := uploadAssets(db, record, "")
		So(err, ShouldNotBeNil)
	})
}

func TestSaveRecord(t *testing.T) {
	Convey("Normal Record", t, func() {
		db := fake.NewFakeDatabase()

		data := map[string]interface{}{"_id": "test/1234", "field1": "str1", "field2": 2}
		record, _ := skyrecord.MakeRecord(data)

		expectedRecord := &skyrecord.Record{
			RecordID: "test/1234",
			Data:     map[string]interface{}{"field1": "str1", "field2": 2, "_reserved": "reserved"},
		}

		err := saveRecord(db, record, "")
		So(err, ShouldBeNil)

		So(db.RecordList["test"]["1234"], ShouldResemble, expectedRecord)
	})

	Convey("Record with wrong ID", t, func() {
		db := fake.NewFakeDatabase()

		data := map[string]interface{}{"_id": "test", "field1": "str1", "field2": 2}
		record, _ := skyrecord.MakeRecord(data)

		err := saveRecord(db, record, "")
		So(err, ShouldNotBeNil)
	})

	Convey("Record with reserved key", t, func() {
		db := fake.NewFakeDatabase()

		data := map[string]interface{}{"_id": "test/1234", "_field1": "str1", "field2": 2}
		record, _ := skyrecord.MakeRecord(data)

		err := saveRecord(db, record, "")
		So(err, ShouldNotBeNil)
	})

	Convey("Record with complex value", t, func() {
		db := fake.NewFakeDatabase()

		data := map[string]interface{}{"_id": "test/1234", "field1": "@str:somestr", "field2": 2}
		record, _ := skyrecord.MakeRecord(data)

		err := saveRecord(db, record, "")
		So(err, ShouldBeNil)

		output, ok := db.RecordList["test"]["1234"]
		So(ok, ShouldBeTrue)

		expectedData := map[string]interface{}{"$type": "str", "$str": "somestr"}
		resultData, ok := output.Data["field1"]
		So(ok, ShouldBeTrue)
		So(resultData, ShouldResemble, expectedData)

		delete(output.Data, "field1")
		expectedRecord := &skyrecord.Record{
			RecordID: "test/1234",
			Data:     map[string]interface{}{"field2": 2, "_reserved": "reserved"},
		}

		So(output, ShouldResemble, expectedRecord)
	})
}

func TestFetchRecord(t *testing.T) {
	Convey("Normal Record", t, func() {
		db := fake.NewFakeDatabase()

		data := map[string]interface{}{"_id": "test/1234", "field1": "str1", "field2": 2}
		record, _ := skyrecord.MakeRecord(data)

		expectedRecord := &skyrecord.Record{
			RecordID: "test/1234",
			Data:     map[string]interface{}{"field1": "str1", "field2": 2},
		}

		err := saveRecord(db, record, "")
		So(err, ShouldBeNil)

		output, err := fetchRecord(db, "test/1234")
		So(err, ShouldBeNil)
		So(output, ShouldResemble, expectedRecord)
	})

	Convey("Record with wrong ID", t, func() {
		db := fake.NewFakeDatabase()

		_, err := fetchRecord(db, "wrongid")
		So(err, ShouldNotBeNil)
	})

	Convey("Record not exist", t, func() {
		db := fake.NewFakeDatabase()

		_, err := fetchRecord(db, "not/exist")
		So(err, ShouldNotBeNil)
	})
}

func TestQueryRecord(t *testing.T) {
	Convey("Normal Record", t, func() {
		db := fake.NewFakeDatabase()

		data1 := map[string]interface{}{"_id": "test/1234", "field1": "str1", "field2": 2}
		record1, _ := skyrecord.MakeRecord(data1)

		data2 := map[string]interface{}{"_id": "test/6789", "field3": "str3", "field4": 4}
		record2, _ := skyrecord.MakeRecord(data2)

		expectedRecord1 := &skyrecord.Record{
			RecordID: "test/1234",
			Data:     map[string]interface{}{"field1": "str1", "field2": 2},
		}
		expectedRecord2 := &skyrecord.Record{
			RecordID: "test/6789",
			Data:     map[string]interface{}{"field3": "str3", "field4": 4},
		}

		expectedList := [2]*skyrecord.Record{expectedRecord1, expectedRecord2}

		err := saveRecord(db, record1, "")
		So(err, ShouldBeNil)
		err = saveRecord(db, record2, "")
		So(err, ShouldBeNil)

		output, err := queryRecord(db, "test")
		So(err, ShouldBeNil)
		So(len(output), ShouldEqual, 2)

		for _, record := range expectedList {
			exist := false

			for _, result := range output {
				if reflect.DeepEqual(record, result) {
					exist = true
				}
			}
			So(exist, ShouldBeTrue)
		}
	})

	Convey("Record type not exist", t, func() {
		db := fake.NewFakeDatabase()

		_, err := queryRecord(db, "notexist")
		So(err, ShouldBeNil)
	})
}
