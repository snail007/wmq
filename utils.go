package main

import (
	"fmt"
	"io/ioutil"
)

//fileGetContents
func fileGetContents(file string) (content string, err error) {
	defer func(err *error) {
		e := recover()
		if e != nil {
			*err = fmt.Errorf("%s", e)
		}
	}(&err)
	bytes, err := ioutil.ReadFile(file)
	content = string(bytes)
	return
}

//value
func value(v interface{}, defaultValue interface{}) interface{} {
	if v != nil {
		return v
	}
	return defaultValue
}
