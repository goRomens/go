package main

import "encoding/json"


///error 为空表示正常 ，如果出错  Error 为错误信息 ，Data 为nil
func ReturnClient(error string,data interface{}) string {
	result:=make(map[string] interface{})
	result["data"]=data
	result["error"]=error
	jsonData, _ := json.Marshal(result)
	return string(jsonData)
}

