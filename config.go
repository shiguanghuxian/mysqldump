package mysqldump

// Config 导出sql所需的配置信息
type Config struct {
	Debug          bool      // 是否调试模式
	IsExportData   bool      // 是否导出数据
	ExportDataStep int64     // 导出数据时，每次查询数据量
	IsCreateDB     bool      // 是否生成建库语句
	OutPath        string    // 输出sql文件目录-绝对路径-用于导出
	SQLPath        string    // 导入的sql文件-绝对路径-用于导入
	OutZip         bool      // 是否导出zip压缩文件
	DbCfg          *DbConfig // 数据库连接信息
}

// DbConfig 数据库连接配置
type DbConfig struct {
	Address string // 数据库连接地址
	Port    int    // 数据库端口
	User    string // 数据库用户名
	Passwd  string // 数据库密码
	DbName  string // 数据库名
}
