
package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"slgserver/config"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var MasterDB *xorm.Engine

var dns string

func init() {
	mysqlConfig, err := config.File.GetSection("mysql")
	if err != nil {
		panic("get mysql config error:")
	}

	fillDns(mysqlConfig)

	// 启动时就打开数据库连接
	if err = initEngine(); err != nil {
		panic(err)
	}

	// 测试数据库连接是否 OK
	if err = MasterDB.Ping(); err != nil {
		panic(err)
	}
}

var (
	ConnectDBErr = errors.New("connect db error")
	UseDBErr     = errors.New("use db error")
)

// TestDB 测试数据库
func TestDB() error {
	mysqlConfig, err := config.File.GetSection("mysql")
	if err != nil {
		fmt.Println("get mysql config error:", err)
		return err
	}

	tmpDns := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local",
		mysqlConfig["user"],
		mysqlConfig["password"],
		mysqlConfig["host"],
		mysqlConfig["port"],
		mysqlConfig["charset"])
	egnine, err := xorm.NewEngine("mysql", tmpDns)
	if err != nil {
		fmt.Println("new engine error:", err)
		return err
	}
	defer egnine.Close()

	// 测试数据库连接是否 OK
	if err = egnine.Ping(); err != nil {
		fmt.Println("ping db error:", err)
		return ConnectDBErr
	}

	_, err = egnine.Exec("use " + mysqlConfig["dbname"])
	if err != nil {
		fmt.Println("use db error:", err)
		_, err = egnine.Exec("CREATE DATABASE " + mysqlConfig["dbname"] + " DEFAULT CHARACTER SET " + mysqlConfig["charset"])
		if err != nil {
			fmt.Println("create database error:", err)

			return UseDBErr
		}

		fmt.Println("create database successfully!")
	}

	// 初始化 MasterDB
	return Init()
}

func Init() error {
	mysqlConfig, err := config.File.GetSection("mysql")
	if err != nil {
		fmt.Println("get mysql config error:", err)
		return err
	}

	fillDns(mysqlConfig)

	// 启动时就打开数据库连接
	if err = initEngine(); err != nil {
		fmt.Println("mysql is not open:", err)
		return err
	}

	return nil
}

func fillDns(mysqlConfig map[string]string) {
	dns = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig["user"],
		mysqlConfig["password"],
		mysqlConfig["host"],
		mysqlConfig["port"],
		mysqlConfig["dbname"],
		mysqlConfig["charset"])

	fmt.Println(dns)
}

func initEngine() error {
	var err error

	MasterDB, err = xorm.NewEngine("mysql", dns)
	if err != nil {
		return err
	}

	maxIdle := config.File.MustInt("mysql", "max_idle", 2)
	maxConn := config.File.MustInt("mysql", "max_conn", 10)

	MasterDB.SetMaxIdleConns(maxIdle)
	MasterDB.SetMaxOpenConns(maxConn)

	showSQL := config.File.MustBool("xorm", "show_sql", false)
	logLevel := config.File.MustInt("xorm", "log_level", 1)

	MasterDB.ShowSQL(showSQL)
	MasterDB.Logger().SetLevel(log.LogLevel(logLevel))

	return nil
}

func StdMasterDB() *sql.DB {
	return MasterDB.DB().DB
}
