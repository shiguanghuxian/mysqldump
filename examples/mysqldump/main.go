package main

import (
	"log"
	"runtime"

	"github.com/shiguanghuxian/mysqldump"
)

func main() {
	// 全部核心运行程序
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 系统日志显示文件和行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	cfg := &mysqldump.Config{
		Debug:        true,
		IsExportData: true,
		IsCreateDB:   false,
		OutPath:      "/Users/zuo/gocode/src/github.com/shiguanghuxian/mysqldump/examples/mysqldump/out/",
		SQLPath:      "/Users/zuo/gocode/src/github.com/shiguanghuxian/mysqldump/examples/mysqldump/out/tslc_test_20180209T084241.sql",
		DbCfg: &mysqldump.DbConfig{
			Address: "127.0.0.1",
			Port:    9031,
			User:    "username",
			Passwd:  "password",
			DbName:  "tslc_test123",
		},
	}
	dm, err := mysqldump.New(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	// 导出
	// dm.Export()
	// 导入
	dm.Import()
}
