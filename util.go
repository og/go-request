package greq

import (
	"fmt"
	ge "github.com/og/x/error"
	core_ogjson "github.com/og/x/json/core"
	"reflect"
)

func structToMap (structValue interface{}, tag string) map[string][]string {
	data := map[string][]string{}
	bList, err := core_ogjson.Marshal(structValue, tag) ;ge.Check(err)
	headerMap := map[string]interface{}{}
	err = core_ogjson.Unmarshal(bList, &headerMap, "header") ; ge.Check(err)
	for key, value := range headerMap {
		rValue := reflect.ValueOf(value)
		switch rValue.Type().Kind() {
		case reflect.String:
			target := value.(string)
			data[key] = append(data[key], target)
		case reflect.Slice:
			for i:=0;i<rValue.Len();i++ {
				itemValue := rValue.Index(i)
				data[key] = append(data[key], fmt.Sprintf(`%v`, itemValue.Interface()))
			}
		default:
			data[key] = append(data[key], fmt.Sprintf(`%v`, value))
		}
	}
	return data
}

