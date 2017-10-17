package main

import (
	"strconv"
	//"database/sql/driver"
	"github.com/satori/go.uuid"

	"strings"


)

//定义路由器结构类型
type Routers struct {
}


func (this *Routers) GetConfig(data interface{},Userinfo interface{}) string  {
    chanCount:=4
    chandata:=make(chan interface{},chanCount)
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
    sqlstring:="select Guid,Name from Bill where CompanyNo='"+CompanyNo+"' and Guid='"+BillGuid+"'"

    sqlbillheader:="select a.Guid,a.BindColumnName,a.IsShow,a.ColumnTitle,a.Width from BillTemplateControls a INNER join BillTemplateDataSource b on b.Guid = a.BillTemplateDataSourceGuid "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='header' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	sqlbillbody:="select a.Guid,a.BindColumnName,a.IsShow,a.ColumnTitle,a.Width from BillTemplateControls a INNER join BillTemplateDataSource b on b.Guid = a.BillTemplateDataSourceGuid "+
		"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='body' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	sqlheaderButtonstring:="SELECT a.ButtonName,a.ButtonExec,a.Status,a.OrderIndex from BillButton a   WHERE a.BillGuid ='"+BillGuid+"' order by a.OrderIndex"

	go getJSON(chandata,sqlstring,"bill")
	go getJSON(chandata,sqlbillheader,"header")
	go getJSON(chandata,sqlbillbody,"body")
	go getJSON(chandata,sqlheaderButtonstring,"button")

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

		return ReturnClient("读取配置错误",nil)
    }else {

		return  ReturnClient("",resultdata)
	}

}


func getSql(BillGuid string) map[string]interface{}  {

	headerSql:="select  b.SqlString,b.GuidKey,b.TableName from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='header' and b.Type='list' and c.BillGuid='"+BillGuid+"'"
	bodySql:="select  b.SqlString,b.GuidKey,b.TableName,b.RelationHeaderKey from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='body' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	chanCount:=2
	chandata:=make(chan interface{},chanCount)

	go getJSON(chandata,headerSql,"header")
	go getJSON(chandata,bodySql,"body")

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


func (this*Routers) GetHeaderData(data interface{},Userinfo interface{}) string {

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	filter:=data.(map[string]interface{})["filter"].(string)
	resultData:=make(map[string]interface{})
    headerconfig:=getSql(BillGuid)
    if headerconfig==nil{
	   	return ReturnClient("读取表头数据源",nil)
	}
	//这里表头只有一行
	header:=headerconfig["header"].([]map[string]interface{})[0]
	//目前这里表体只支持一个 后期需要支持多个的话 需要修改这里
	body:=headerconfig["body"].([]map[string]interface{})[0]
	resultData["headerconfig"]=header
	resultData["bodyconfig"]=body

	headerSql:=header["SqlString"].(string)

	SetSqlString(Userinfo.(map[string]interface{}),&headerSql)


	//目前这里没有加 参数过滤，后期加的话 要放这里
	SetSqlStringForHeader(header,&headerSql,filter)


		headerRows:=GetDataForOneTable(headerSql)
		resultData["headerrow"]=headerRows
		if headerRows==nil{
			return ReturnClient("表头数据获取失败",nil)
		}

	return ReturnClient("",resultData)
}


func (this*Routers) GetBodyData(data interface{},Userinfo interface{}) string {

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)

	HeaderMainGuid:=data.(map[string]interface{})["HeaderMainGuid"].(string)

	config:=getSql(BillGuid)
	if config==nil{
		return ReturnClient("读取配置信息失败",nil)
	}

	//目前这里表体只支持一个 后期需要支持多个的话 需要修改这里
	body:=config["body"].([]map[string]interface{})[0]

	bodySql:=body["SqlString"].(string)

	SetSqlString(Userinfo.(map[string]interface{}),&bodySql)

	SetSqlStringForBody(body,&bodySql)


	GetBodyForDataString(HeaderMainGuid,body,&bodySql)
	bodyRows:=GetDataForOneTable(bodySql)

	if bodyRows==nil{
		return ReturnClient("表体数据获取失败",nil)
	}

	return ReturnClient("",bodyRows)
}



func (this *Routers) GetEditConfig(data interface{},Userinfo interface{}) string  {
	chanCount:=2
	chandata:=make(chan interface{},chanCount)

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	GuidValue:=data.(map[string]interface{})["GuidValue"].(string)
	DetailGuidValue:=data.(map[string]interface{})["DetailGuidValue"].(string)
	Type:=data.(map[string]interface{})["Type"].(string)


	resultArray:= make(map[string]interface{})

	sqlheaderconfig:="select  b.TableName,b.SQLString,b.RelationHeaderKey,b.GuidKey,b.TableAnotherName from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='"+Type+"' and b.Type='edit' and c.BillGuid='"+BillGuid+"'"

	sqlheaderContols:="select   a.ControlType,a.ColumnTitle Title,a.BindColumnName,a.IsSave,a.IsShow,a.DefaultValue,a.DataSelectRelation,a.DataSelectRelationDetail from BillTemplateControls a  inner join BillTemplateDataSource b  on a.BillTemplateDataSourceGuid=b.Guid "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='"+Type+"' and b.Type='edit' and c.BillGuid='"+BillGuid+"' ORDER BY a.OrderIndex"


	go getJSON(chandata,sqlheaderconfig,"config")
	go getJSON(chandata,sqlheaderContols,"controls")


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
    	 return ReturnClient("读取编辑配置错误",nil)
	}

	SqlString:=resultdata["config"].([]map[string]interface{})[0]["SQLString"].(string)
	SetSqlString(Userinfo.(map[string]interface{}),&SqlString)

	GuidKey:=resultdata["config"].([]map[string]interface{})[0]["GuidKey"].(string)
	TableAnotherName:=resultdata["config"].([]map[string]interface{})[0]["TableAnotherName"].(string)

	//如果修改的是表体 则GuidValue 就是表体Guid
	if(Type=="body"){
		GetEditSqlString(GuidKey,DetailGuidValue,&SqlString,TableAnotherName)
	}else {
		GetEditSqlString(GuidKey,GuidValue,&SqlString,TableAnotherName)
	}


	headerchan:=make(chan interface{})
	go getJSON(headerchan,SqlString,"row")
	headerRow:=(<-headerchan).(map[string]interface{})["table"]


	headerRow=GetDefaultRow(resultdata["controls"].([]map[string]interface{}),headerRow.([]map[string]interface{}),Userinfo.(map[string]interface{}),GuidValue)

	resultArray["controls"]=resultdata["controls"]
	resultArray["row"]=headerRow.([]map[string]interface{})[0]

	return ReturnClient("",resultArray)
}

func  (this *Routers) InsertHeader(data interface{},Userinfo interface{}) string  {
	//chanCount:=2
	//chandata:=make(chan interface{},chanCount)

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	Data:=data.(map[string]interface{})["Data"].([]interface{})

	sqlheaderconfig:="select TableName,GuidKey from  BillEditHeader where BillGuid='"+BillGuid+"'"

	headerchan:=make(chan interface{})
	go getJSON(headerchan,sqlheaderconfig,"row")

	headerconfig:=(<-headerchan).(map[string]interface{})["table"]
	TableName:=headerconfig.([]map[string]interface{})[0]["TableName"].(string)
	GuidKey:=headerconfig.([]map[string]interface{})[0]["GuidKey"].(string)
	GuidValue:=""
	for _,item:=range Data {
	   if item.(map[string]interface{})["BindColumnName"].(string)==	GuidKey{
		   GuidValue=item.(map[string]interface{})["value"].(string)
	   }
	}
	sqlString:=GetInsertHeaderForSql(TableName,GuidValue,GuidKey,Data)

	headerchan=make(chan interface{})
	go exec(headerchan,sqlString,"row")

	execresult:=(<-headerchan).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("新增表头失败",nil)
}

func (this*Routers)DeleteHeaderData(data interface{},Userinfo interface{}) string {
	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	GuidValue:=data.(map[string]interface{})["GuidValue"].(string)

	sqlheaderconfig:="select  b.GuidKey,b.TableName from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='header' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	sqlbodyconfig:="select  b.RelationHeaderKey,b.TableName from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='body' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	chanCount:=2
	chandata:=make(chan interface{},chanCount)

	go getJSON(chandata,sqlheaderconfig,"headerconfig")
	go getJSON(chandata,sqlbodyconfig,"bodyconfig")


	resultdata := make(map[string]interface{})

	for i:=0;i<chanCount;i++  {
		result:=<-chandata
		resultdata[result.(map[string]interface{})["chanName"].(string)]=(result.(map[string]interface{})["table"].([]map[string]interface{})[0])

	}
	deletesqls:= GetdeleteForSqlString(GuidValue,resultdata["headerconfig"].(map[string]interface{}),resultdata["bodyconfig"].(map[string]interface{}))

	headerchan:=make(chan interface{})
	go exec(headerchan,deletesqls,"row")

	execresult:=(<-headerchan).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("删除失败",nil)
}


func (this*Routers)InsertData(data interface{},Userinfo interface{}) string {

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	Data:=data.(map[string]interface{})["Data"].([]interface{})
	Type:=data.(map[string]interface{})["Type"].(string)


	sqlconfig:="select   b.TableName,b.GuidKey from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='"+Type+"' and b.Type='edit' and c.BillGuid='"+BillGuid+"'"

	headerconfig:=GetTableWithChan(sqlconfig)

	TableName:=headerconfig[0]["TableName"].(string)
	GuidKey:=headerconfig[0]["GuidKey"].(string)

	GuidValue:=""
	for _,item:=range Data {
		if item.(map[string]interface{})["BindColumnName"].(string)==GuidKey{
			GuidValue=item.(map[string]interface{})["value"].(string)
		}
	}
	sqlString:=GetInsertHeaderForSql(TableName,GuidValue,GuidKey,Data)

	resutlt:= ExecSqls(sqlString)
	return ReturnClient(resutlt,nil)
}

func (this*Routers)DeleteBodeyData(data interface{},Userinfo interface{})  string {

	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	GuidValue:=data.(map[string]interface{})["GuidValue"].(string)

	sqlheaderconfig:="select  b.GuidKey,b.TableName from BillTemplateDataSource b "+
	"INNER JOIN BillTemplate c on b.BillTemplateGuid=c.Guid where c.BillType='body' and b.Type='list' and c.BillGuid='"+BillGuid+"'"

	chandata:=make(chan interface{})
	go getJSON(chandata,sqlheaderconfig,"bodyconfig")

	result:=<-chandata
	resultdata:=(result.(map[string]interface{})["table"].([]map[string]interface{})[0])


	deletesqls:= GetdeleteForBodySqlString(GuidValue,resultdata)

	headerchan:=make(chan interface{})
	go exec(headerchan,deletesqls,"row")

	execresult:=(<-headerchan).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("删除失败",nil)
}

func (this*Routers)GetDataSelect(data interface{},Userinfo interface{}) string {
	DataselectType:=data.(map[string]interface{})["DataselectType"].(string)
	Name:=""
	if data.(map[string]interface{})["Name"]!=nil{
		Name=data.(map[string]interface{})["Name"].(string)
	}
	Page:=int(data.(map[string]interface{})["Page"].(float64))
	Count:=int(data.(map[string]interface{})["Count"].(float64))

	resaultdata:=make(map[string]interface{})

	sql:="select SqlString,GuidKey,NameKey,HiddenColumn from BillDataSelect where DataSelectType='"+DataselectType+"'"
	resultdata:=GetTableWithChan(sql)

    sqlString:=resultdata[0]["SqlString"].(string)
	GuidKey:=resultdata[0]["GuidKey"].(string)
	NameKey:=resultdata[0]["NameKey"].(string)
	HiddenColumn:=resultdata[0]["HiddenColumn"].(string)


	resaultdata["HiddenColumn"]=HiddenColumn

	SetSqlString(Userinfo.(map[string]interface{}),&sqlString)
	//1替换选择条件 组合SQL
	sql= GetDataSeletSql(GuidKey,NameKey,Name,sqlString)
	//2根据SQL字符串 获取总共有多少行的数据
    sqlRowCount:=GetDataSeletSqlForRowsCount(sql)
    //3最后在替换分页
    sql=GetDataSeletSqlForlimit(sql,strconv.Itoa(Page),strconv.Itoa(Count))


	chandata:=make(chan interface{},2)
	go getJSON(chandata,sql,"row")
	go getJSON(chandata,sqlRowCount,"rowcount")


	chanresultdata := make(map[string]interface{})

	for i:=0;i<2;i++  {
		result:=<-chandata
		chanresultdata[result.(map[string]interface{})["chanName"].(string)]=result

	}

	DataTable:=(chanresultdata["row"]).(map[string]interface{})
	DataTableRowCount:=((chanresultdata["rowcount"]).(map[string]interface{})["table"].([]map[string]interface{})[0])
	resaultdata["Columns"]=DataTable["columns"]
	resaultdata["RowsCount"]=DataTableRowCount["RowCount"]
	resaultdata["Rows"]=DataTable["table"].([]map[string]interface{})

	return ReturnClient("",resaultdata)
}

func (this*Routers)GetBillMenu(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)

	sql:="SELECT NAME,BillGuid,IndexNo from BillMenu where CompanyNo='"+CompanyNo+"'"
	chandata:=make(chan interface{})
	go getJSON(chandata,sql,"menu")
	row:=(<-chandata).(map[string]interface{})["table"].([]map[string]interface{})
	return ReturnClient("",row)
}

func (this*Routers)GetBill(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	sql:="select  IFNULL(a.Guid,'') Guid,IFnull(a.BillType,'') BillType,Ifnull(a.Name,'') Name,b.Guid BillGuid,b.Name BillName from Bill b left join BillTemplate a on a.BillGuid=b.Guid  where b.CompanyNo='"+CompanyNo+"'  ORDER BY b.Name "
	chandata:=make(chan interface{})
	go getJSON(chandata,sql,"menu")
	row:=(<-chandata).(map[string]interface{})["table"].([]map[string]interface{})
	return ReturnClient("",row)
}

func (this*Routers)GetBillData(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	sql:="SELECT Guid,Name from Bill WHERE CompanyNo='"+CompanyNo+"'"
	chandata:=make(chan interface{})
	go getJSON(chandata,sql,"menu")
	row:=(<-chandata).(map[string]interface{})["table"].([]map[string]interface{})
	return ReturnClient("",row)
}



func (this*Routers)InsertBill(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	Name:=data.(map[string]interface{})["Name"].(string)
	Guid:=data.(map[string]interface{})["Guid"].(string)
	if len(Guid)==0{
		Guid=uuid.NewV4().String()
	}
	sqls:=make([]string,0)
	sqls=append(sqls,"DELETE from Bill where Guid='"+Guid+"'")
	sqls=append(sqls,"DELETE from BillMenu where BillGuid='"+Guid+"'")
	sqls=append(sqls,"insert into Bill(Guid,Name,CompanyNo) values('"+Guid+"','"+Name+"','"+CompanyNo+"')")
	sqls=append(sqls,"insert into BillMenu (Guid,Name,BillGuid,CompanyNo) values('"+Guid+"','"+Name+"','"+Guid+"','"+CompanyNo+"')")


	chandata:=make(chan interface{})
	go exec(chandata,sqls,"bill")
	execresult:=(<-chandata).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("操作失败",nil)
}



func (this*Routers)DeleteBillData(data interface{},Userinfo interface{}) string {

	Guid:=data.(map[string]interface{})["Guid"].(string)

	sqls:=make([]string,0)
	sqls=append(sqls,"DELETE from Bill where Guid='"+Guid+"'")
	sqls=append(sqls,"DELETE from BillMenu where BillGuid='"+Guid+"'")

	chandata:=make(chan interface{})
	go exec(chandata,sqls,"bill")
	execresult:=(<-chandata).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("操作失败",nil)
}



func (this*Routers)InsertBillTemplate(data interface{},Userinfo interface{}) string {
	Name:=data.(map[string]interface{})["Name"].(string)
	BillGuid:=data.(map[string]interface{})["BillGuid"].(string)
	BillType:=data.(map[string]interface{})["BillType"].(string)
	Guid:=data.(map[string]interface{})["Guid"].(string)
	if len(Guid)==0{
		Guid=uuid.NewV4().String()
	}
	sqls:=make([]string,0)
	sqls=append(sqls,"DELETE from BillTemplate where Guid='"+Guid+"'")
	sqls=append(sqls,"insert into BillTemplate (Guid,Name,BillGuid,BillType) values('"+Guid+"','"+Name+"','"+BillGuid+"','"+BillType+"')")


	chandata:=make(chan interface{})
	go exec(chandata,sqls,"bill")
	execresult:=(<-chandata).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("操作失败",nil)
}


func (this*Routers)DeleteBillTemplate(data interface{},Userinfo interface{}) string {

	Guid:=data.(map[string]interface{})["Guid"].(string)

	sqls:=make([]string,0)
	sqls=append(sqls,"DELETE from BillTemplate where Guid='"+Guid+"'")


	chandata:=make(chan interface{})
	go exec(chandata,sqls,"bill")
	execresult:=(<-chandata).(map[string]interface{})

	if execresult["error"].(string)==""{
		return ReturnClient("","")
	}
	return ReturnClient("操作失败",nil)
}



func (this*Routers)GetBillTemplateDataSource(data interface{},Userinfo interface{}) string {
	BillTemplateGuid:=data.(map[string]interface{})["BillTemplateGuid"].(string)

	sql:="select Guid,BillTemplateGuid,SqlString,TableName,RelationHeaderKey,GuidKey,Type,TableAnotherName from BillTemplateDataSource where BillTemplateGuid ='"+BillTemplateGuid+"'"

	sqls:=make([]map[string]string,0)
	sqls=append(sqls,map[string]string{"BillTemplateDataSource":sql})


	DataTable:=GetTables(sqls)
    if DataTable==nil{
		ReturnClient("获取数据失败",nil)
	}
	return ReturnClient("",DataTable)
}



func (this*Routers)InsertBillDataSource(data interface{},Userinfo interface{}) string {
	datamap:=data.(map[string]interface{})["Data"].(map[string]interface{})

	Guid:=datamap["Guid"].(string)
	if len(Guid)==0{
			Guid=uuid.NewV4().String()
	}
	BillTemplateGuid:=datamap["BillTemplateGuid"].(string)
	SqlString:=datamap["SqlString"].(string)
	TableName:=datamap["TableName"].(string)
	RelationHeaderKey:=datamap["RelationHeaderKey"].(string)
	GuidKey:=datamap["GuidKey"].(string)
	Type:=datamap["Type"].(string)
	TableAnotherName:=datamap["TableAnotherName"].(string)

	sqldelete:="delete from BillTemplateDataSource where Guid='"+Guid+"'"

	sqlInsert:="Insert into BillTemplateDataSource(Guid,BillTemplateGuid,SqlString,TableName,RelationHeaderKey,GuidKey,Type,TableAnotherName) values ("+
 "'"+Guid+"','"+BillTemplateGuid+"','"+SqlString+"','"+TableName+"','"+RelationHeaderKey+"','"+GuidKey+"','"+Type+"','"+TableAnotherName+"')"

	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)
	sqls=append(sqls,sqlInsert)

	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}


func (this*Routers)DeleteBillDataSource(data interface{},Userinfo interface{}) string {
	Guid:=data.(map[string]interface{})["Guid"].(string)


	sqldelete:="delete from BillTemplateDataSource where Guid='"+Guid+"'"

	sqlInsert:="delete from BillTemplateControls WHERE BillTemplateDataSourceGuid='"+Guid+"'"

	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)
	sqls=append(sqls,sqlInsert)

	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}



func (this*Routers)GetBillTemplateControlsForSqlColumns(data interface{},Userinfo interface{}) string {
	BillTemplateDataSourceGuid:=data.(map[string]interface{})["BillTemplateDataSourceGuid"].(string)

sqlsource:="select SqlString from BillTemplateDataSource where guid='"+BillTemplateDataSourceGuid+"'"
chane:=make(chan interface{})
go getJSON(chane,sqlsource,"SqlString")
sqlstring:=(<-chane).(map[string]interface{})["table"].([]map[string]interface{})[0]["SqlString"].(string)
sqlstring= strings.Replace(sqlstring, "[选择条件]", " and 1=2 ", -1)


chane=make(chan interface{})
go getJSON(chane,sqlstring,"SqlString")
columns:=(<-chane).(map[string]interface{})["columns"]

return ReturnClient("",columns)
}



func (this*Routers)GetBillTemplateControls(data interface{},Userinfo interface{}) string {
	BillTemplateDataSourceGuid:=data.(map[string]interface{})["BillTemplateDataSourceGuid"].(string)

	sql:="select Guid,BillTemplateDataSourceGuid,ControlType,ColumnTitle,BindColumnName,IsSave,IsShow,OrderIndex,DefaultValue,DataSelectRelation,DataSelectRelationDetail from BillTemplateControls where BillTemplateDataSourceGuid='"+BillTemplateDataSourceGuid+"'"

	sqls:=make([]map[string]string,0)
	sqls=append(sqls,map[string]string{"BillTemplateDataSource":sql})


	DataTable:=GetTables(sqls)
	if DataTable==nil{
		ReturnClient("获取数据失败",nil)
	}

	resultdata:=DataTable.(map[string]interface{})["BillTemplateDataSource"]
	return ReturnClient("",resultdata)
}

func (this*Routers)GetBillTemplateControlsForDataSelect(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	Name:=data.(map[string]interface{})["Name"].(string)

	sql:="select Name,DataSelectType from BillDataSelect where CompanyNo='"+CompanyNo+"' and Name like '%"+Name+"%'"

	sqls:=make([]map[string]string,0)
	sqls=append(sqls,map[string]string{"BillDataSelect":sql})

	DataTable:=GetTables(sqls)
	if DataTable==nil{
		ReturnClient("获取数据失败",nil)
	}

	resultdata:=DataTable.(map[string]interface{})["BillDataSelect"]
	return ReturnClient("",resultdata)
}

func (this*Routers)GetBillTemplateControlsForDataSelectColumns(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)
	DataSelectType:=data.(map[string]interface{})["DataSelectType"].(string)

	sql:="select SqlString from BillDataSelect where CompanyNo='"+CompanyNo+"' and DataSelectType='"+DataSelectType+"'"

	sqls:=make([]map[string]string,0)
	sqls=append(sqls,map[string]string{"BillDataSelect":sql})

	DataTable:=GetTables(sqls)
	if DataTable==nil{
		ReturnClient("获取数据失败",nil)
	}

	resultdata:=DataTable.(map[string]interface{})["BillDataSelect"].([]map[string]interface{})
	if len(DataSelectType)==0{
		return ReturnClient("请配置标准查询","")
	}

	SqlString:=resultdata[0]["SqlString"].(string)
	SqlString=strings.Replace(SqlString,"[选择条件]"," and 1=2 ",-1)
	result:= GetTableColumnsAndRows(SqlString)
	if result==nil{
		ReturnClient("标准查询获取SQl出错了","")
	}
	return ReturnClient("",result.(map[string]interface{})["columns"])
}



func (this*Routers)InsertBillTemplateControls(data interface{},Userinfo interface{}) string {
	datamap:=data.(map[string]interface{})["Data"].(map[string]interface{})

	Guid:=datamap["Guid"].(string)
	if len(Guid)==0{
		Guid=uuid.NewV4().String()
	}

	BillTemplateDataSourceGuid:=datamap["BillTemplateDataSourceGuid"].(string)
	ControlType:=datamap["ControlType"].(string)
	ColumnTitle:=datamap["ColumnTitle"].(string)
	BindColumnName:=datamap["BindColumnName"].(string)
	IsSave:=datamap["IsSave"].(string)
	IsShow:=datamap["IsShow"].(string)

	OrderIndex:=strconv.FormatFloat(datamap["OrderIndex"].(float64), 'f', -1, 64)
	DefaultValue:=datamap["DefaultValue"].(string)
	DataSelectRelation:=datamap["DataSelectRelation"].(string)
	DataSelectRelationDetail:=datamap["DataSelectRelationDetail"].(string)

	sqldelete:="delete from BillTemplateControls where Guid='"+Guid+"'"

	sqlInsert:="Insert into BillTemplateControls(Guid,BillTemplateDataSourceGuid,ControlType,ColumnTitle,BindColumnName,IsSave,IsShow,OrderIndex,DefaultValue,DataSelectRelation,DataSelectRelationDetail) values ("+
		"'"+Guid+"','"+BillTemplateDataSourceGuid+"','"+ControlType+"','"+ColumnTitle+"','"+BindColumnName+"','"+IsSave+"','"+IsShow+"','"+OrderIndex+"','"+DefaultValue+"','"+DataSelectRelation+"','"+DataSelectRelationDetail+"')"

	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)
	sqls=append(sqls,sqlInsert)

	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}


func (this*Routers)DeleteBillTemplateControls(data interface{},Userinfo interface{}) string {
	Guid:=data.(map[string]interface{})["Guid"].(string)


	sqldelete:="delete from BillTemplateControls where Guid='"+Guid+"'"


	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)


	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}


func (this*Routers)GetBillDataSelect(data interface{},Userinfo interface{}) string {
	CompanyNo:=Userinfo.(map[string]interface{})["CompanyNo"].(string)

	sql:="select Guid,Name,SqlString,GuidKey,NameKey,HiddenColumn,DataSelectType,CompanyNo from BillDataSelect where CompanyNo='"+CompanyNo+"' "

	sqls:=make([]map[string]string,0)
	sqls=append(sqls,map[string]string{"BillTemplateDataSource":sql})


	DataTable:=GetTables(sqls)
	if DataTable==nil{
		ReturnClient("获取数据失败",nil)
	}

	resultdata:=DataTable.(map[string]interface{})["BillTemplateDataSource"]
	return ReturnClient("",resultdata)
}


func (this*Routers)InsertBillDataSelect(data interface{},Userinfo interface{}) string {
	CompanyNo := Userinfo.(map[string]interface{})["CompanyNo"].(string)
	datamap := data.(map[string]interface{})["Data"].(map[string]interface{})
	statue := "修改"
	Guid := datamap["Guid"].(string)
	if len(Guid) == 0 {
		statue = "新增"
		Guid = uuid.NewV4().String()
	}

	//Name,SqlString,GuidKey,NameKey,HiddenColumn,DataSelectType,CompanyNo
	Name := datamap["Name"].(string)
	SqlString := datamap["SqlString"].(string)
	GuidKey := datamap["GuidKey"].(string)
	NameKey := datamap["NameKey"].(string)
	HiddenColumn := datamap["HiddenColumn"].(string)
	DataSelectType := datamap["DataSelectType"].(string)

	if statue == "新增" {
		execsql := "select count(*) num from BillDataSelect where CompanyNo='" + CompanyNo + "' and DataSelectType='" + DataSelectType + "'"
		count := GetTableWithChan(execsql)
		counts := count[0]["num"].(int64)
		if counts > 0 {
			return ReturnClient("类型出现重复，请修改", "")
		}
	}
	sqldelete:="delete from BillDataSelect where Guid='"+Guid+"'"

	sqlInsert:="Insert into BillDataSelect(Guid,Name,SqlString,GuidKey,NameKey,HiddenColumn,DataSelectType,CompanyNo) values ("+
		"'"+Guid+"','"+Name+"','"+SqlString+"','"+GuidKey+"','"+NameKey+"','"+HiddenColumn+"','"+DataSelectType+"','"+CompanyNo+"')"

	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)
	sqls=append(sqls,sqlInsert)

	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}


func (this*Routers)DeleteBillDataSelect(data interface{},Userinfo interface{}) string {
	Guid:=data.(map[string]interface{})["Guid"].(string)


	sqldelete:="delete from BillDataSelect where Guid='"+Guid+"'"


	sqls:=make([]string,0)
	sqls=append(sqls,sqldelete)


	result:=ExecSqls(sqls)
	if len(result)==0{
		return ReturnClient("",result)
	}
	return ReturnClient(result,"")
}
