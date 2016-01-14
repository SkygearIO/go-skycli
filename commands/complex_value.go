package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var ComplexTypeList []ComplexType

type ComplexType interface {
	Validate(string) bool
	Convert(string) (string, error)
}

// Location
type ComplexLocation struct {
	validRegexp *regexp.Regexp
}

func NewComplexLocation() *ComplexLocation {
	return &ComplexLocation{
		validRegexp: regexp.MustCompile("^@loc:"),
	}
}

func (s *ComplexLocation) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *ComplexLocation) Convert(valStr string) (string, error) {
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

	loc := map[string]interface{}{"$type": "geo", "$lat": resultVal[0], "$lng": resultVal[1]}
	locJson, err := json.Marshal(loc)
	if err != nil {
		return "", err
	}

	return string(locJson), nil
}

// Reference
type ComplexReference struct {
	validRegexp *regexp.Regexp
}

func NewComplexReference() *ComplexReference {
	return &ComplexReference{
		validRegexp: regexp.MustCompile("^@ref:"),
	}
}

func (s *ComplexReference) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *ComplexReference) Convert(valStr string) (string, error) {
	str := s.validRegexp.ReplaceAllString(valStr, "")

	ref := map[string]interface{}{"$type": "ref", "$id": str}
	refStr, err := json.Marshal(ref)
	if err != nil {
		return "", err
	}
	return string(refStr), nil
}

// String
type ComplexString struct {
	validRegexp *regexp.Regexp
}

func NewComplexString() *ComplexString {
	return &ComplexString{
		validRegexp: regexp.MustCompile("^@str:"),
	}
}

func (s *ComplexString) Validate(valStr string) bool {
	return s.validRegexp.MatchString(valStr)
}

func (s *ComplexString) Convert(valStr string) (string, error) {
	str := s.validRegexp.ReplaceAllString(valStr, "")

	strMap := map[string]interface{}{"$type": "str", "$str": str}
	strStr, err := json.Marshal(strMap)
	if err != nil {
		return "", err
	}
	return string(strStr), nil
}

func init() {
	ComplexTypeList = append(ComplexTypeList, NewComplexLocation())
	ComplexTypeList = append(ComplexTypeList, NewComplexReference())
	ComplexTypeList = append(ComplexTypeList, NewComplexString())
}
