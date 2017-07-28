package cache

import "reflect"

func isPointer(a interface{}) bool {
	kind := reflect.TypeOf(a).Kind()
	return kind == reflect.Ptr || kind == reflect.UnsafePointer
}
