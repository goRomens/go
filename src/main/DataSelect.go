package main

import "strings"

func GetDataSeletSql(GuidKey string,NameKey string,Name string,sqlString string) string  {
	filter:=""

		if len(Name) != 0 {
			filter = " and " + NameKey + " like '" + Name + "%'"
		}
	sql:= strings.Replace(sqlString, "[选择条件]", filter, -1)
	return sql
}


func GetDataSeletSqlForRowsCount(sqlString string) string  {
	filter:="select COUNT(*) RowCount "
	index:= strings.Index(strings.ToLower(sqlString),"from")
	sql := sqlString[index :]
	filter+=sql
	return filter
}

func GetDataSeletSqlForlimit(sqlString string,Page string,Count string) string  {
	sqlString+=" limit "+Page+","+Count+""
	return sqlString
}