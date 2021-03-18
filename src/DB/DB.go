package db

import (
	"database/sql"
	"fmt"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"test/src/Common"
)

type MysqlConfig struct {
	userName		string
	password  		string
	ip 				string
	port 			string
	dbName 			string
}

type ChatDB struct {
	MysqlConfig
	db *sql.DB
}

func InitDB() *ChatDB {
	chatDB := ChatDB{}
	config := MysqlConfig{
		userName: "root",
		password: "Sx86427353",
		ip:       "127.0.0.1",
		port:     "3306",
		dbName:   "chat_db",
	}
	chatDB.MysqlConfig = config
	chatDB.Connect()

	//初始化需要用到的ID序列
	chatDB.InitUserID()
	return &chatDB
}

func (d * ChatDB) InitUserID() {
	NowMinUserID := d.GetMaxUserID()
	var iniUserID uint64 = Common.MaxUInt64(NowMinUserID,100000)
	Common.SetUserID(iniUserID)
}

var ChatDBInstance *ChatDB

func GetDBInstance() *ChatDB {
	if ChatDBInstance == nil {
		db := InitDB()
		ChatDBInstance = db
	}
	return ChatDBInstance
}

func (d *ChatDB) GetMaxUserID() uint64 {
	var minUserID uint64
	err := d.db.QueryRow("SELECT  MAX(userID) FROM g_user").Scan(
		&minUserID)
	if err != nil {
		fmt.Println("GetMinUserID wrong",err)
		return 100000
	}
	return minUserID
}

func (d *ChatDB) Connect() {
	path := strings.Join([]string{d.userName, ":",
		d.password, "@tcp(", d.ip, ":", d.port, ")/", d.dbName, "?charset=utf8"}, "")
	d.db, _ = sql.Open("mysql", path)
	d.db.SetMaxOpenConns(10)
	d.db.SetMaxIdleConns(10)
	//验证连接
	if err := d.db.Ping(); err != nil{
		fmt.Println("open database fail")
		panic("open database fail")
	}
	fmt.Println("mysql connect success")
}

func (d *ChatDB) Insert(table string,colNameArray []string, args ...interface{}) bool {
	//开启事务
	tx, err := d.db.Begin()
	if err != nil{
		fmt.Println("tx fail")
		return false
	}
	//准备sql语句
	colNameSqlStr := "("
	colWSqlStr := "("
	for _,colName := range colNameArray {
		colNameSqlStr += " `" + colName + "`,"
		colWSqlStr += "?,"
	}
	colNameSqlStr = colNameSqlStr[:len(colNameSqlStr) - 1]
	colWSqlStr = colWSqlStr[:len(colWSqlStr) - 1]
	colNameSqlStr += ")"
	colWSqlStr += ")"
	ResultSql := "INSERT INTO " + table + colNameSqlStr + " VALUES " + colWSqlStr
	insertSql, err := tx.Prepare(ResultSql)
	if err != nil{
		fmt.Println("Prepare fail")
		return false
	}
	//将参数传递到sql语句中并且执行
	_, err = insertSql.Exec(args...)
	if err != nil{
		fmt.Println("Exec fail",err)
		return false
	}
	//将事务提交
	tx.Commit()
	return true
}

func (d *ChatDB) Delete(table string,deleteKeyName string,i interface{}) bool {
	tx, err := d.db.Begin()
	if err != nil{
		fmt.Println("tx fail")
	}
	//准备sql语句
	stmt, err := tx.Prepare("DELETE FROM " + table +" WHERE " + deleteKeyName +" = ?")
	if err != nil{
		fmt.Println("Prepare fail")
		return false
	}
	//设置参数以及执行sql语句
	_, err = stmt.Exec(i)
	if err != nil{
		fmt.Println("Exec fail")
		return false
	}
	//提交事务
	tx.Commit()
	return true
}

func (d *ChatDB) Update(table string,colNameArray []string, args ...interface{}) bool{
	//开启事务
	tx, err := d.db.Begin()
	if err != nil{
		fmt.Println("tx fail")
		return false
	}
	//准备sql语句
	colNameSqlStr := ""
	colWSqlStr := ""
	colNameArrayLen := len(colNameArray)
	for i,colName := range colNameArray {
		if i != colNameArrayLen - 1{
			colNameSqlStr += colName + " = ?,"
		}else{
			colWSqlStr = colName + " = ?"
		}
	}
	colNameSqlStr = colNameSqlStr[:len(colNameSqlStr) - 1]
	colWSqlStr = colWSqlStr[:len(colWSqlStr) - 1]
	insertSql, err := tx.Prepare("UPDATE " + table + " SET " + colNameSqlStr + " WHERE " + colWSqlStr)
	if err != nil{
		fmt.Println("Prepare fail")
		return false
	}
	//将参数传递到sql语句中并且执行
	_, err = insertSql.Exec(args)
	if err != nil{
		fmt.Println("Exec fail")
		return false
	}
	//将事务提交
	tx.Commit()
	return true
}