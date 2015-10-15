package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"
	"unsafe"
)

type CommndLineArguments struct {
	Configfile string `required:"false" name:"config" default:"/etc/daemon.conf" description:"Конфигурационный	файл"`
	Daemon     bool   `required:"true" name:"daemon" default:"false" description:"Запуск	приложения	в	режиме	daemon"`
}

var (
	f *flag.FlagSet
)

func GetArguments(variable interface{}) (err error) {
	variableType := reflect.TypeOf(variable)

	if variableType.Kind() != reflect.Ptr {
		return errors.New("Variable is not a pointer ")
	}

	val := reflect.ValueOf(variable).Elem()

	f = flag.NewFlagSet("default", flag.ContinueOnError)

	numFields := val.NumField()

	requiredFieldsMap := make(map[string]bool)

	for i := 0; i < numFields; i++ {

		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		varPointer := (unsafe.Pointer(valueField.Addr().Pointer()))

		desc := tag.Get("description")
		name := tag.Get("name")

		if utf8.RuneCountInString(desc) == 0 {
			return errors.New("Empty description")
		}

		switch valueField.Interface().(type) {
		case bool:
			defaultVal, _ := strconv.ParseBool(tag.Get("default"))
			f.BoolVar((*bool)(varPointer), name, defaultVal, desc)
		case time.Duration:
			defaultVal, _ := time.ParseDuration(tag.Get("default"))
			f.DurationVar((*time.Duration)(varPointer), name, defaultVal, desc)
		case float64:
			defaultVal, _ := strconv.ParseFloat(tag.Get("default"), 64)
			f.Float64Var((*float64)(varPointer), name, defaultVal, desc)
		case int:
			defaultVal, _ := strconv.ParseInt(tag.Get("default"), 10, 32)
			flag.IntVar((*int)(varPointer), name, int(defaultVal), desc)
		case int64:
			defaultVal, _ := strconv.ParseInt(tag.Get("default"), 10, 64)
			f.Int64Var((*int64)(varPointer), name, defaultVal, desc)
		case uint:
			defaultVal, _ := strconv.ParseUint(tag.Get("default"), 10, 32)
			f.UintVar((*uint)(varPointer), name, uint(defaultVal), desc)
		case uint64:
			defaultVal, _ := strconv.ParseUint(tag.Get("default"), 10, 64)
			f.Uint64Var((*uint64)(varPointer), name, defaultVal, desc)
		case string:
			f.StringVar((*string)(varPointer), name, tag.Get("default"), desc)
		default:
			return errors.New("Field type is not supported")
		}

		isRequired, _ := strconv.ParseBool(tag.Get("required"))
		if isRequired {
			requiredFieldsMap[name] = true
		}

	}

	err = f.Parse(os.Args[1:])

	if err == nil {

		f.Visit(
			func(flag *flag.Flag) {
				delete(requiredFieldsMap, flag.Name)
			})

		if len(requiredFieldsMap) != 0 {
			keys := make([]string, 0, len(requiredFieldsMap))
			for k := range requiredFieldsMap {
				keys = append(keys, k)
			}
			err = errors.New(fmt.Sprintf("Required keys not found: %s", keys))
		}

	}

	return err

}

func main() {

	var Commands *CommndLineArguments = new(CommndLineArguments)

	if err := GetArguments(Commands); err != nil {
		fmt.Printf("%v", err)
		f.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Configuration file: %s\n", Commands.Configfile)
	fmt.Printf("Daemon: %v\n", Commands.Daemon)

}
