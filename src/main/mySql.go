package main

import (
	_ "github.com/go-sql-driver/mysql"

)

func exec(result chan interface{},sqlStrings []string,chanName string)  {
	resultData:=make(map[string]interface{})

	tx,err := db.Begin()
	if err!=nil{
		resultData["chanName"]=chanName
		resultData["error"]="事务出错了"
		result <- resultData
		return
	}

	for _,sqlstring:=range sqlStrings {
		stmt, err := tx.Prepare(sqlstring)
		if err != nil {
			tx.Rollback()
			resultData["chanName"]=chanName
			resultData["error"]="Prepare出错了"
			result <- resultData
			return
		}
		defer stmt.Close()
		_, err2 := stmt.Exec()
		if err2 != nil {
			tx.Rollback()
			resultData["chanName"]=chanName
			resultData["error"]="执行出错了"
			result <- resultData
			return
		}
	}
	tx.Commit()
	resultData["chanName"]=chanName
	resultData["error"]=""
	result <- resultData

}


func getJSON(result chan interface{},sqlString string,chanName string)  {
	resultData:=make(map[string]interface{})
	stmt, err := db.Prepare(sqlString)
	if err != nil {
		resultData["chanName"]=chanName
		resultData["error"]="Prepare出错了"
		result <- resultData
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		resultData["chanName"]=chanName
		resultData["error"]="查询出错了"
		result <- resultData
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()

	if err != nil {
		resultData["chanName"]=chanName
		resultData["error"]="获取列出错了"
		result <- resultData
		return
	}

	tableData := make([]map[string]interface{}, 0)

	count := len(columns)
	tableColumns:=make([]interface{},0)

	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for _, col := range columns {
		columnclass:= make(map[string]interface{})
		columnclass["Name"]=col
		tableColumns=append(tableColumns,columnclass)

	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			resultData["chanName"]=chanName
			resultData["error"]="赋值出错了"
			result <- resultData
			return
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			b, ok := v.([]byte)
			if (ok) {
				entry[col] = string(b)
			} else {
				entry[col] = v
			}
		}

		tableData = append(tableData, entry)
	}
	resultData["chanName"]=chanName
	resultData["error"]=""
	resultData["table"]=tableData
	resultData["columns"]=tableColumns
	result <- resultData

}


func GetTables(sqls []map[string]string) interface{}  {
	chanCount:=len(sqls)
	chandata:=make(chan interface{},chanCount)

	for _,col := range sqls{
		for k, v := range col {
			go getJSON(chandata,v,k)
		}
	}


	resultdata := make(map[string]interface{})

	isHasError:=false
	for i:=0;i<chanCount;i++  {
		result:=<-chandata
		err:=result.(map[string]interface{})["error"].(string)
		if err !=""{
			isHasError=true
		} else {
			resultdata[result.(map[string]interface{})["chanName"].(string)]=result.(map[string]interface{})["table"]
		}
	}
	if isHasError{
		return  nil
	}
	return resultdata
}

func GetTablesColumnsAndRows(sqls []map[string]string) interface{}  {
	chanCount:=len(sqls)
	chandata:=make(chan interface{},chanCount)

	for _,col := range sqls{
		for k, v := range col {
			go getJSON(chandata,v,k)
		}
	}


	resultdata := make(map[string]interface{})

	isHasError:=false
	for i:=0;i<chanCount;i++  {
		result:=<-chandata
		err:=result.(map[string]interface{})["error"].(string)
		if err !=""{
			isHasError=true
		} else {
			resultdata[result.(map[string]interface{})["chanName"].(string)]=result.(map[string]interface{})
		}
	}
	if isHasError{
		return  nil
	}
	return resultdata
}

func GetTableColumnsAndRows(sql string) interface{}  {
	result:= GetTablesColumnsAndRows([]map[string]string{map[string]string{"name":sql}})
	if result==nil{
		return  nil
	}

	return result.(map[string]interface{})["name"]
}

func GetTableWithChan(sql string) []map[string]interface{} {
	result:=make(chan interface{})
	go getJSON(result,sql,"row")
	return (<-result).(map[string]interface{})["table"].([]map[string]interface{})
}

func ExecSqls(sqls[]string)string  {
	headerchan:=make(chan interface{})
	go exec(headerchan,sqls,"row")
	execresult:=(<-headerchan).(map[string]interface{})
	return execresult["error"].(string)
}