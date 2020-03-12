package sample

import (
	"errors"
	"reflect"
	"regexp"
)

func (s Sample) Check(input interface{}) (bool, error) {
	switch s.Spec.Type {
	case Row:
		return reflect.DeepEqual(s.Rows, input), nil
	case Regex:
		str := input.(string)
		return regexp.MatchString(s.Regex, str)
	case SPJ:
		return false, errors.New("not implemented")
	case SQL:
		return false, errors.New("not implemented")
	default:
		return false, errors.New("unable to tell the sample type")
	}
}
