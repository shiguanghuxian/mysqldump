package mysqldump

/* 导出数据时使用到的模型 */

// CreateTable 创建table sql查询
type CreateTable struct {
	Table       string `xorm:"'Table'"`
	CreateTable string `xorm:"'Create Table'"`
}

// CreateDb 创建数据库 sql查询
type CreateDb struct {
	Database       string `xorm:"'Database'"`
	CreateDatabase string `xorm:"'Create Database'"`
}

// TPLSqlModel 导出数据sql部分
type TPLSqlModel struct {
	TableName string // 表名
	CreateSQL string // 创建表sql语句
	InsertSQL string // 插入数据sql语句
}

// TPLModel 导出数据结构体
type TPLModel struct {
	MySQL *DbConfig
	SQL   []*TPLSqlModel
	Date  string
}

// TableColumn 用于读取数据库每一列数据
type TableColumn interface {
}
