// backend/database.go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB 初始化数据库连接和创建表
func InitDB() error {
	var err error
	db, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试数据库连接
	if err = db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 创建用户表
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	// 创建系统设置表
	createSettingsTableSQL := `
	CREATE TABLE IF NOT EXISTS system_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		setting_key TEXT UNIQUE NOT NULL,
		setting_value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createSettingsTableSQL)
	if err != nil {
		return fmt.Errorf("创建系统设置表失败: %v", err)
	}

	log.Println("数据库初始化成功")
	return nil
}

// CreateUser 创建新用户
func CreateUser(username, passwordHash string, isAdmin bool) error {
	query := `INSERT INTO users (username, password_hash, is_admin) VALUES (?, ?, ?)`
	_, err := db.Exec(query, username, passwordHash, isAdmin)
	if err != nil {
		return fmt.Errorf("创建用户失败: %v", err)
	}
	return nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password_hash, is_admin FROM users WHERE username = ?`
	row := db.QueryRow(query, username)

	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return user, nil
}

// IsSystemInitialized 检查系统是否已经初始化
func IsSystemInitialized() (bool, error) {
	query := `SELECT setting_value FROM system_settings WHERE setting_key = 'initialized'`
	var value string
	err := db.QueryRow(query).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value == "true", nil
}

// SetSystemInitialized 设置系统为已初始化
func SetSystemInitialized() error {
	query := `INSERT OR REPLACE INTO system_settings (setting_key, setting_value, updated_at) VALUES ('initialized', 'true', CURRENT_TIMESTAMP)`
	_, err := db.Exec(query)
	return err
}

// HasAnyUsers 检查是否有任何用户
func HasAnyUsers() (bool, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
