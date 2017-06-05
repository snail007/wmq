package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	logger "github.com/snail007/mini-logger"
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
func inArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}
func pathExists(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func ctxFunc(fname string) logger.MiniLogger {
	return log.With(logger.Fields{"func": fname})
}
