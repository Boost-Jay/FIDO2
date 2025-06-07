package db

import (
	"database/sql"
	"fido2/config"
	"fido2/internal/entity"
	"fido2/pkg/utils"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dbContext struct {
	mu sync.Mutex
	db *gorm.DB
}

var (
	instance *dbContext
	once     sync.Once
)

func GetDB() *gorm.DB {
	if instance == nil {
		Connect()
	}
	instance.mu.Lock()
	defer instance.mu.Unlock()
	return instance.db
}

func Connect() {
	once.Do(func() {
		host := config.GetEnv("DB_HOST")
		port := config.GetEnv("DB_PORT")
		user := config.GetEnv("DB_USER")
		password := config.GetEnv("DB_PASSWORD")
		dbName := config.GetEnv("DB_NAME")
		sslmode := config.GetEnv("DB_SSLMODE")

		// 1. 先連到 postgres 預設庫 (postgres)，檢查/建立目標 DB (webauthn)
		defaultDSN := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
			host, port, user, password, sslmode,
		)
		sqlDB, err := sql.Open("postgres", defaultDSN)
		if err != nil {
			panic(fmt.Sprintf("failed to open default postgres db: %v", err))
		}
		defer sqlDB.Close()

		// 重試 Ping 機制，避免 PG 尚在啟動
		for i := 0; i < 5; i++ {
			if err = sqlDB.Ping(); err == nil {
				break
			}
			time.Sleep(time.Second * 1)
		}
		if err != nil {
			panic(fmt.Sprintf("cannot ping default postgres db: %v", err))
		}

		// 檢查目標資料庫是否存在
		var exists bool
		checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
		if err = sqlDB.QueryRow(checkSQL).Scan(&exists); err != nil {
			panic(fmt.Sprintf("failed to check db existence: %v", err))
		}
		if !exists {
			// 資料庫不存在就建立
			createSQL := fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)
			if _, err := sqlDB.Exec(createSQL); err != nil {
				panic(fmt.Sprintf("failed to create database %s: %v", dbName, err))
			}
			utils.GetLogger().Infof("Database %s created successfully", dbName)
		} else {
			utils.GetLogger().Infof("Database %s already exists", dbName)
		}

		// 2. 用 GORM 連到目標資料庫
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Taipei",
			host, port, user, password, dbName, sslmode,
		)

		var gormDB *gorm.DB
		for i := 0; i < 5; i++ {
			gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err == nil {
				break
			}
			time.Sleep(time.Second * 1)
		}
		if err != nil {
			panic(fmt.Sprintf("failed to connect database after retries: %v", err))
		}

		// 3. 只 AutoMigrate User 這一張表
		if err := gormDB.AutoMigrate(&entity.User{}); err != nil {
			utils.GetLogger().Fatalf("failed to auto migrate: %v", err)
		}
		utils.GetLogger().Info("User table migrated successfully")

		instance = &dbContext{db: gormDB}
	})
}