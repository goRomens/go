package main

import "strings"

func SetSqlStringForHeader(userinfo map[string]interface{},sqlstring *string,filter string)  {
	*sqlstring=strings.Replace(*sqlstring, "[选择条件]", filter, -1)
}

func SetSqlStringForBody(bodyconfig map[string]interface{},sqlstring *string)  {

}

func GetDataForOneTable(sqlstring string) []map[string]interface{}   {
	chandata:=make(chan interface{})
	go getJSON(chandata,sqlstring,"header")
	headerRows:=<-chandata

	if headerRows.(map[string]interface{})["error"].(string)==""{
		return headerRows.(map[string]interface{})["table"].([]map[string]interface{})
	} else {
		return nil
	}
}

//替换全局参数的
func SetSqlString(userinfo map[string]interface{},sqlstring *string)  {
	CompanyNo:=userinfo["CompanyNo"].(string)
	*sqlstring = strings.Replace(*sqlstring, "[企业号]", CompanyNo, -1)

}

//获取表头 连接表体是的 关键连接字段所对应的值 目前默认取第一行的值
func GetHeaderMainKeyforValue(headerconfig map[string]interface{},headRow []map[string] interface{}) string  {
	GuidKey:=headerconfig["GuidKey"].(string)
	if len(headRow)==0{
		return ""
	}
   return 	headRow[0][GuidKey].(string)
}

//获取最终表体数据
func GetBodyForDataString(headerGuid string,bodyconfig map[string]interface{},bodySql*string)   {
	RelationHeaderKey:=bodyconfig["RelationHeaderKey"].(string)
	filterstring:=" and "+RelationHeaderKey+" = '"+headerGuid+"'"
	*bodySql=strings.Replace(*bodySql, "[选择条件]", filterstring, -1)
}


