package pkg

import (
	"fmt"
	"log"
	"time"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/Kaikai20040827/graduation/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func NewAdminDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/mysql?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.AdminUser,
        cfg.AdminPassword,
        cfg.Host,
        cfg.Port,
    )

    return gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
}

func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
    //数据库登入基本信息
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
    fmt.Println("✓ Database login basic information done")

    //配置访问时的日志模式
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    }
    fmt.Println("✓ Database logger mode configuration done")

    //创建一个数据库访问实例
    db, err := gorm.Open(mysql.Open(dsn), gormConfig)
    if err != nil {
        return nil, err
    }
    fmt.Println("✓ Creating a database connection done")

    //连接数据库
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    fmt.Println("✓ Connecting the database done")

    //配置连接
    sqlDB.SetConnMaxIdleTime(20 * time.Minute) //最大闲置连接时间
    sqlDB.SetConnMaxLifetime(time.Hour) //最大连接生命周期: Hour（小时）
    sqlDB.SetMaxOpenConns(100) //最大打开的连接数量
    fmt.Println("✓ Setting configuration of the connection done")

    //全局保存这个数据库访问
    DB = db

    if err := migrate(db); err != nil {
		log.Printf("自动迁移失败: %v", err)
	}
    fmt.Println("✓ Migrating the database done")

    //无异常，则返回数据库
    return db, nil
}

func migrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &model.User{},
        &model.File{},
    )
}
