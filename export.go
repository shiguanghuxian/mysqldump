package mysqldump

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/mholt/archiver"
)

// Export 导出数据库所有表
func (md *Mysqldump) Export() (outFile string, err error) {
	if md.isClose == true {
		return "", errors.New("已调用Close关闭相关资源，无法进行导出")
	}
	// 创建导出sql文件
	outFile = fmt.Sprintf("%s/%s_%s.sql", strings.TrimRight(md.cfg.OutPath, "/"), md.cfg.DbCfg.DbName, time.Now().Format("20060102T150405"))
	lf, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return "", err
	}
	defer func() {
		// 关闭文件
		lf.Close()
		// 压缩文件
		if md.cfg.OutZip == true {
			outZipFile := fmt.Sprintf("%s/%s_%s.zip", strings.TrimRight(md.cfg.OutPath, "/"), md.cfg.DbCfg.DbName, time.Now().Format("20060102T150405"))
			err = archiver.Zip.Make(outZipFile, []string{outFile})
			if err == nil {
				outFile = outZipFile
			}
		}
	}()

	// 获取建库语句
	createSQL, err := md.GetCreateDbSQL()
	if err != nil {
		return "", err
	}
	// 获取数据库字符集
	charSet := "utf8"
	var valid = regexp.MustCompile("CHARACTER SET ([a-z0-9A-Z]+) ")
	finds := valid.FindAllStringSubmatch(createSQL, -1)
	if len(finds) > 0 {
		if len(finds[0]) > 1 {
			charSet = finds[0][1]
		}
	}
	// 判断是否需要建库语句
	if md.cfg.IsCreateDB == true {
		createSQL = fmt.Sprintf("CREATE DATABASE `%s`; /*!40100 DEFAULT CHARACTER SET %s */", md.cfg.DbCfg.DbName, charSet)
	} else {
		createSQL = ""
	}

	// 写入头部信息
	_, err = lf.WriteString(fmt.Sprintf(`/*
		中防电信sql导出
	   
		数据库地址        : %s:%d
		数据库类型        : MySQL
		数据库名         : %s
	   
		生成时间: %s
*/

%s

SET NAMES %s;
SET FOREIGN_KEY_CHECKS = 0;

`,
		md.cfg.DbCfg.Address,
		md.cfg.DbCfg.Port,
		md.cfg.DbCfg.DbName,
		time.Now().Format("2006-01-02 15:04:05"),
		createSQL,
		charSet))
	if err != nil {
		return "", err
	}

	// 查询数据库表列表
	tables, err := md.SelectTableNames()
	if err != nil {
		return "", err
	}

	// 导出数据对象
	tplSqlModel := make([]*TPLSqlModel, 0)
	// 循环表名，查询出对应的表创建语句
	for _, table := range tables {
		log.Println("导出", table)
		// 导出建表语句
		sql, err := md.GetCreateTableSQL(table)
		if err != nil {
			return "", err
		}
		tplSqlModel = append(tplSqlModel, &TPLSqlModel{
			TableName: table,
			CreateSQL: sql,
		})
		log.Println(sql)
		// 写入一个表到建表语句
		_, err = lf.WriteString(fmt.Sprintf(
			`%s-- ----------------------------
-- Table structure for %s
-- ----------------------------
DROP TABLE IF EXISTS %s%s%s;
%s;
%s`,
			"\n\n",
			table,
			"`",
			table,
			"`",
			sql,
			"\n"))
		if err != nil {
			return "", err
		}
		// 导出数据
		if md.cfg.IsExportData == true {
			md.ExportData(lf, table)
		}
	}
	// js, _ := json.Marshal(aa)
	// log.Println(string(js))
	return outFile, nil
}

// SelectTableNames 查询数据库表列表
func (md *Mysqldump) SelectTableNames() (tables []string, err error) {
	tables = make([]string, 0)
	err = md.conn.SQL("SHOW TABLES;").Cols(fmt.Sprintf("Tables_in_%s", md.cfg.DbCfg.DbName)).Find(&tables)
	return
}

// GetCreateTableSQL 查询创建表语句
func (md *Mysqldump) GetCreateTableSQL(tableName string) (string, error) {
	creates := make([]*CreateTable, 0)
	err := md.conn.SQL(fmt.Sprintf("show create table %s", tableName)).Find(&creates)
	log.Println(err)
	if err != nil {
		return "", err
	}
	if len(creates) == 0 {
		return "", errors.New("查询table 创建语句错误")
	}
	return creates[0].CreateTable, nil
}

// GetCreateDbSQL 获取创建数据库
func (md *Mysqldump) GetCreateDbSQL() (string, error) {
	createSQLs := make([]*CreateDb, 0)
	err := md.conn.SQL(fmt.Sprintf("SHOW CREATE DATABASE %s", md.cfg.DbCfg.DbName)).Find(&createSQLs)
	if err != nil {
		return "", err
	}
	if len(createSQLs) == 0 {
		return "", errors.New("查询创建数据库语句为空")
	}
	return createSQLs[0].CreateDatabase, nil
}

// ExportData 导出数据为
func (md *Mysqldump) ExportData(w io.Writer, tableName string) (err error) {
	log.Println("开始导出数据:", tableName)
	// 查询总数据行数
	var count int64
	count, err = md.conn.Table(tableName).Count()
	if err != nil {
		return
	}
	log.Println(count)

	columns, xormColumns, err := md.conn.Dialect().GetColumns(tableName)
	if err != nil {
		return err
	}

	var offset int64
	for offset = 0; offset < count; offset += md.cfg.ExportDataStep {
		colNames := md.conn.Dialect().Quote(strings.Join(columns, md.conn.Dialect().Quote(", ")))
		sql := fmt.Sprintf("select %s from %s limit %d offset %d", colNames, tableName, md.cfg.ExportDataStep, offset)
		list, err := md.conn.QueryInterface(sql)
		if err != nil {
			return err
		}
		for _, one := range list {
			// 拼接插入语句头部
			installSQL := fmt.Sprintf("\nINSERT INTO %s (%s) VALUES ",
				md.conn.Dialect().Quote(tableName),
				colNames)
			values := make([]string, 0)
			for _, column := range columns {
				val, ok := one[column] // 读取本行值
				if ok == false {
					return errors.New("列名和值无法对应")
				}
				// 判断是否是时间类型
				if xormColumn, ok := xormColumns[column]; ok == true {
					aa[xormColumn.SQLType.Name] = xormColumn.SQLType.Name
					if xormColumn.SQLType.IsTime() == true {
						isTimeNull := false
						if val == nil {
							val = "null"
							isTimeNull = true
						} else {
							valTime := val.(time.Time)
							// valTime, err := time.ParseInLocation("2006-01-02T15:04:05Z", val.(string), time.UTC)
							if err == nil {
								val = valTime.Format("2006-01-02 15:04:05")
							} else {
								val = "null"
								isTimeNull = true
							}
						}
						if isTimeNull == true {
							values = append(values, fmt.Sprintf("%v", val))
						} else {
							values = append(values, fmt.Sprintf("'%v'", val))
						}

					} else if xormColumn.SQLType.IsBlob() == true {
						if val == nil {
							val = "false"
						} else {
							if reflect.TypeOf(val).Kind() == reflect.Slice {
								val = md.conn.Dialect().FormatBytes(val.([]byte))
							} else if reflect.TypeOf(val).Kind() == reflect.String {
								val = val.(string)
							}
						}

						values = append(values, fmt.Sprintf("%v", val))
					} else if xormColumn.SQLType.IsNumeric() == true {
						if val == nil {
							val = "null"
						} else {
							if valByte, ok := val.([]byte); ok == true {
								// log.Println(column, "3-1")
								val = string(valByte)
							} else {
								// log.Println(column, "3-2")
								val = fmt.Sprint(val)
							}
						}

						values = append(values, fmt.Sprintf("%v", val))
					} else {
						if val == nil {
							val = ""
						} else {
							if valByte, ok := val.([]byte); ok == true {
								// log.Println(column, "3-1")
								val = string(valByte)
							} else {
								// log.Println(column, "3-2")
								val = fmt.Sprint(val)
							}
						}

						values = append(values, fmt.Sprintf("'%v'", val))
					}
				}

			}
			// 拼接插入语句值部分
			installSQL = fmt.Sprintf("%s (%s);", installSQL, strings.Join(values, ","))
			// 写入数据
			_, err = io.WriteString(w, installSQL)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

var aa map[string]string

func init() {
	aa = make(map[string]string, 0)
}
