package util

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func JsonDump(i interface{}) string {
	js, _ := json.Marshal(i)
	return string(js)
}
func Env(key string, def string) string {
	v, exist := os.LookupEnv(key)
	if !exist {
		v = def
	}
	return v
}
func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
func Fallback(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
func MD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
