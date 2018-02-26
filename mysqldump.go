package mysqldump

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

// Mysqldump mysql导出数据对象
type Mysqldump struct {
	conn    *xorm.Engine
	cfg     *Config
	isClose bool
}

// New 创建一个Mysqldump对象
func New(cfg *Config) (*Mysqldump, error) {
	if cfg == nil {
		return nil, errors.New("配置信息不能为nil")
	}
	// 处理配置信息
	if cfg.OutPath == "" {
		return nil, errors.New("导出sql输出路径不能是空")
	}
	if cfg.ExportDataStep == 0 {
		cfg.ExportDataStep = 1000
	}
	// 创建导出对象
	mysqldump := &Mysqldump{
		cfg:     cfg,
		isClose: false,
	}
	// 连接mysql
	err := mysqldump.OpenMysql()
	if err != nil {
		return nil, err
	}

	return mysqldump, nil
}

// Close 不使用导出功能时，关闭连接资源
func (md *Mysqldump) Close() error {
	md.isClose = true
	return md.conn.Close()
}

// OpenMysql 连接mysql
func (md *Mysqldump) OpenMysql() error {
	// 拼接连接数据库字符串
	connStr := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=UTC",
		md.cfg.DbCfg.User,
		md.cfg.DbCfg.Passwd,
		md.cfg.DbCfg.Address,
		md.cfg.DbCfg.Port,
		md.cfg.DbCfg.DbName)

	// 连接数据库
	engine, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		return err
	}

	// 是否开启debug模式
	if md.cfg.Debug {
		engine.Logger().SetLevel(core.LOG_DEBUG) // 调试信息
		engine.ShowSQL(true)                     // 显示sql
	}
	engine.SetMaxIdleConns(2)            // 空闲连接池数量
	engine.SetMaxOpenConns(8)            // 最大连接数
	engine.SetMapper(core.GonicMapper{}) // 命名规则

	// 设置数据库时区
	engine.DatabaseTZ = time.UTC
	engine.TZLocation = time.UTC

	md.conn = engine

	log.Println("连接数据库成功")
	return nil
}

// GetRootDir 获取程序跟目录,返回值尾部包含'/'
func (md *Mysqldump) GetRootDir() string {
	// 文件不存在获取执行路径
	file, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		file = fmt.Sprintf(".%s", string(os.PathSeparator))
	} else {
		file = fmt.Sprintf("%s%s", file, string(os.PathSeparator))
	}
	return file
}
