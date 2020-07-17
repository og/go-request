package greq

import (
	gtest "github.com/og/x/test"
	"os"
	"reflect"
	"testing"
)
type Sub struct {
	SubName string `test:"3"`
}
type Demo struct {
	Name string `test:"1"`
	Sub `test:"2"`
	Sub2 struct{
		Sub2Name string `test:"5"`
	} `test:"4"`
	File *os.File `test:"6"`
	UploadPhoto `test:"7"`
	List []string `test:"9"`
}
type UploadPhoto struct {
	Photo *os.File `test:"8"`
}
func TestScan(t *testing.T) {
	as := gtest.NewAS(t)
	_ = as
	demo := Demo{}
	data := []string{
		"1:Name string",
		"2:Sub greq.Sub",
		"3:SubName string",
		`4:Sub2 struct { Sub2Name string "test:\"5\"" }`,
		"5:Sub2Name string",
		"6:File *os.File",
		"7:UploadPhoto greq.UploadPhoto",
		"8:Photo *os.File",
		"9:List []string",
	}
	actual := []string{}
	scan(demo, func(rValue reflect.Value, field reflect.StructField) {
		actual = append(actual, field.Tag.Get("test")+ ":" + field.Name + " " + field.Type.String())
	})
	as.Equal(data, actual)
}
