// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package create

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (ExecutionConfig) elemValueOrNil(v interface{}) interface{} {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return reflect.Zero(t.Elem()).Interface()
		} else {
			return reflect.ValueOf(v).Interface()
		}
	} else if v == nil {
		return reflect.Zero(t).Interface()
	}

	return v
}

func (ExecutionConfig) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in ExecutionConfig and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg ExecutionConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("ExecutionConfig", pflag.ExitOnError)
	cmdFlags.StringVarP(&(executionConfig.Version),fmt.Sprintf("%v%v", prefix, "version"), "v", "", "version of the launch plan to be executed.")
	cmdFlags.StringVarP(&(executionConfig.File),fmt.Sprintf("%v%v", prefix, "file"), "f", "", "file for the execution params.If not specified defaults to <launchplan_name>.inputs.yaml.")
	cmdFlags.BoolVarP(&(executionConfig.GenParam),fmt.Sprintf("%v%v", prefix, "genParam"), "g", false, "flag to indicate the generation of params file.")
	return cmdFlags
}
