package main
import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"reflect"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"os"
	"io"
)

var db *sql.DB

func initDB() {
	db, _= sql.Open("mysql", "sunbo:sunbo2017@tcp(571b37a3593e8.bj.cdb.myqcloud.com:8400)/CloudBI?charset=utf8");
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()
}
func BIHandle(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			Error.Println(e)
			//for i := 0; i < 10; i++ {
				funcName, file, line, ok := runtime.Caller(4)
				if ok {
					Error.Println("[func:%v,file:%v,line:%v]\n",  runtime.FuncForPC(funcName).Name(), file, line)
				}
			//}

			fmt.Fprintf(w,ReturnClient("服务端出现错误，请联系管理员222",""))
		}
	}()
	result, _:= ioutil.ReadAll(req.Body)
	req.Body.Close()

	var f interface{}
	json.Unmarshal(result, &f)
	m := f.(map[string]interface{})
	var queryType=m["queryType"].(string)

	ruTest:=&Routers{}
	vf := reflect.ValueOf(ruTest)


	//创建带调用方法时需要传入的参数列表
	parms := []reflect.Value{reflect.ValueOf(m["params"]),reflect.ValueOf(m["userInfo"])}
	returndata:=vf.MethodByName(queryType).Call(parms)

	fmt.Fprintf(w, returndata[0].String())
}
var (
   Trace *log.Logger // 记录所有日志
   Info *log.Logger // 重要的信息
   Warning *log.Logger // 需要注意的信息
   Error *log.Logger // 非常严重的问题
  )

func init() {
	 file, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	 if err != nil {
		 log.Fatalln("Failed to open error log file:", err)
		 }


	 Error = log.New(io.MultiWriter(file, os.Stderr),
		 "ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	 }
func main() {

	initDB()
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/", BIHandle)
	err := http.ListenAndServe(":8029", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}

}
