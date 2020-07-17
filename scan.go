package greq

import (
	"reflect"
)
type ScanHandle func(reflect.Value, reflect.StructField)
func coreScan(rValue reflect.Value, handle ScanHandle) {
	rType := rValue.Type()
	for i:=0;i<rValue.NumField();i++ {
		itemValue := rValue.Field(i)
		structField := rType.Field(i)
		handle(itemValue, structField)
		if itemValue.Type().Kind() == reflect.Struct {
			coreScan(itemValue, handle)
		}
	}
}
func scan (v interface{}, handle ScanHandle) {
	rootValue := reflect.ValueOf(v)
	coreScan(rootValue, handle)
}