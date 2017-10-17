package main

import (
	"strings"
	"github.com/satori/go.uuid"
	"fmt"
)

func GetEditSqlString(GuidKey string,GuidValue string,sqlstring *string,TableAnotherName string)  {

	if(len(TableAnotherName)!=0){
		GuidKey=TableAnotherName+"."+GuidKey
	}
	filterstring:=" and "+GuidKey+" = '"+GuidValue+"'"
	*sqlstring = strings.Replace(*sqlstring, "[选择条件]", filterstring, -1)
}


func GetDefaultRow(config []map[string] interface{},row []map[string] interface{},Userinfo map[string] interface{},RelationHeaderKey string)[]map[string] interface{}{
	//CompanyNo:=Userinfo["CompanyNo"].(string)
	//新增行 设置默认值
    if len(row)==0{
    	newrow:=make(map[string]interface{})
		for _,item:=range config {
			newitem:=item

			defaultvalue:=""
			if newitem["DefaultValue"]!=nil{
				defaultvalue=newitem["DefaultValue"].(string)
			}

			setrow(newrow,item["BindColumnName"].(string),"",defaultvalue,Userinfo,RelationHeaderKey)
		}
		row=append(row,newrow)
	//如果是修改行，切没有值时 需要用默认值替换下
	} else
	{
		for _,item:=range config {
			defaultvalue:=""
			if item["DefaultValue"]!=nil{
				defaultvalue=item["DefaultValue"].(string)
			}
	    	rowDefautlvalue:=""
			if(row[0][item["BindColumnName"].(string)]!=nil) {
				rowDefautlvalue=row[0][item["BindColumnName"].(string)].(string)
			}

			if len(defaultvalue)>0&&len(rowDefautlvalue)==0{
				setrow(row[0],item["BindColumnName"].(string),"",defaultvalue,Userinfo,RelationHeaderKey)
			}

		}

	}
  return  row
}

func setrow(row map[string]interface{},key string,value string,defaultvalue string,Userinfo map[string] interface{},RelationHeaderKey string)  {
	CompanyNo:=Userinfo["CompanyNo"].(string)

	row[key]=value
	switch defaultvalue {
	 case"[NEWID]":
		 row[key]=uuid.NewV4().String()
		 fmt.Println(row[key])
	 	break
	case"[企业号]":
		row[key]=CompanyNo
		break
	//这个关键字只要用于表体新增时 赋值表头的主关键字
	case "[RelationHeaderKey]":
		row[key]=RelationHeaderKey
		break

	}
}


func GetInsertHeaderForSql(TableName string,GuidValue string,GuidKey string,Data []interface{}) []string {
	execsql:=make([]string,2)

	execsql[0]="delete from "+TableName+" where "+GuidKey+" ='"+GuidValue+"'"

	sqlinsert:="insert into "+TableName+"("
	columns:=""
	values:=""
	for i,vv:=range Data {
		v:=vv.(map[string]interface{})
		if i==len(Data)-1{
			columns+=v["BindColumnName"].(string)
			values+="'"+v["value"].(string)+"'"
		} else {
		   columns+=v["BindColumnName"].(string)+","
		   values+="'"+v["value"].(string)+"',"
		}

	}
	sqlinsert+=columns+") values ("+values+")"

	execsql[1]=sqlinsert
	return execsql
}

func GetdeleteForSqlString(GuidValue string,headerconfig map[string]interface{},bodyconfig map[string]interface{})[]string {
	sqlstring:=make([]string,2)
	bodyTable:=bodyconfig["TableName"].(string)
	RelationHeaderKey:=bodyconfig["RelationHeaderKey"].(string)

	sqldeleteforbody:="delete from "+bodyTable+" where "+RelationHeaderKey+" = '"+GuidValue+"'"
	sqlstring[0]=sqldeleteforbody

	headerTable:=headerconfig["TableName"].(string)
    GuidKey:=headerconfig["GuidKey"].(string)

	sqldeleteforheader:="delete from "+headerTable+" where "+GuidKey+" = '"+GuidValue+"'"
	sqlstring[1]=sqldeleteforheader
	return sqlstring
}

func GetdeleteForBodySqlString(GuidValue string,headerconfig map[string]interface{}) []string {
	sqlstring:=make([]string,1)
	headerTable:=headerconfig["TableName"].(string)
	GuidKey:=headerconfig["GuidKey"].(string)
	sqldeleteforheader:="delete from "+headerTable+" where "+GuidKey+" = '"+GuidValue+"'"
	sqlstring[0]=sqldeleteforheader
	return sqlstring
}

