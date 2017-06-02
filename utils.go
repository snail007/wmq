package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
func initConfig() (err error) {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.String("api-port", "0.0.0.0:3302", "api service listening port")
	pflag.Parse()
	viper.BindPFlag("listen.api", pflag.Lookup("api-port"))

	viper.SetConfigName("config")     // name of config file (without extension)
	viper.AddConfigPath("/etc/wmq/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.wmq") // call multiple times to add many search paths
	viper.AddConfigPath(".")          // optionally look for config in the working directory
	viper.ReadInConfig()              // Find and read the config file
	fmt.Println(viper.GetString("listen.api"))
	return errors.New("test")
}
