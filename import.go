package mysqldump

import "errors"

// Import 导入sql文件到数据库
func (md *Mysqldump) Import(sqlPath ...string) (err error) {
	if md.isClose == true {
		return errors.New("已调用Close关闭相关资源，无法进行导入")
	}
	if len(sqlPath) > 0 {
		_, err = md.conn.ImportFile(sqlPath[0])
	} else {
		_, err = md.conn.ImportFile(md.cfg.SQLPath)
	}
	return err
}
