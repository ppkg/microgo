package conn

import (
	"log"
	"os"
	"time"

	"github.com/maybgit/glog"
	"github.com/ppkg/microgo/sys"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetDB(dsn string) *gorm.DB {
	// 禁用默认事务
	// 为了确保数据一致性，GORM 会在事务里执行写入操作（创建、更新、删除）。如果没有这方面的要求，您可以在初始化时禁用它，这将获得大约 30%+ 性能提升。
	// 需要用到事务的地方单独处理
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: &gormlogger{
			Writer: log.New(os.Stdout, "\r\n", log.LstdFlags),
			Config: logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		},
	})
	if err != nil {
		glog.Error("打开数据库连接错误：", err)
	}

	sys.UseTrace(func() {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			glog.Error(err)
		}
	})

	sqlDB, err := db.DB()
	if err != nil {
		glog.Error("打开数据库连接错误：", err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	if sys.IsDebug() {
		db = db.Debug()
	}
	return db
}
