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

package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ComplexTypeList provide the list of available complex type
var ComplexTypeList []complexType

type complexType interface {
	Validate(string) bool
	Convert(string) (interface{}, error)
}

// Location
type complexLocation struct {
	validRegexp *regexp.Regexp
}

func newComplexLocation() *complexLocation {
	return &complexLocation{
		validRegexp: regexp.MustCompile("^@loc:"),
	}
}

func (s *complexLocation) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *complexLocation) Convert(valStr string) (interface{}, error) {
	str := s.validRegexp.ReplaceAllString(valStr, "")
	resultStr := strings.Split(str, ",")
	if len(resultStr) != 2 {
		return "", fmt.Errorf("Wrong format of complex value(location).")
	}

	var resultVal []float64
	for _, x := range resultStr {
		rx, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return "", err
		}
		resultVal = append(resultVal, rx)
	}

	loc := map[string]interface{}{
		"$type": "geo",
		"$lat":  resultVal[0],
		"$lng":  resultVal[1],
	}
	return loc, nil
}

// Reference
type complexReference struct {
	validRegexp *regexp.Regexp
}

func newComplexReference() *complexReference {
	return &complexReference{
		validRegexp: regexp.MustCompile("^@ref:"),
	}
}

func (s *complexReference) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *complexReference) Convert(valStr string) (interface{}, error) {
	if s.Validate(valStr) == false {
		return "", fmt.Errorf("Unexpected complex reference")
	}

	str := s.validRegexp.ReplaceAllString(valStr, "")

	ref := map[string]interface{}{"$type": "ref", "$id": str}
	return ref, nil
}

// String
type complexString struct {
	validRegexp *regexp.Regexp
}

func newComplexString() *complexString {
	return &complexString{
		validRegexp: regexp.MustCompile("^@str:"),
	}
}

func (s *complexString) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *complexString) Convert(valStr string) (interface{}, error) {
	if s.Validate(valStr) == false {
		return "", fmt.Errorf("Unexpected complex string")
	}

	str := s.validRegexp.ReplaceAllString(valStr, "")

	strMap := map[string]interface{}{"$type": "str", "$str": str}
	return strMap, nil
}

func init() {
	ComplexTypeList = append(ComplexTypeList, newComplexLocation())
	ComplexTypeList = append(ComplexTypeList, newComplexReference())
	ComplexTypeList = append(ComplexTypeList, newComplexString())
}
