// backend/user.go
package main

// User 用户数据模型
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // 不在JSON中显示密码哈希
	IsAdmin      bool   `json:"is_admin"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// InitRequest 系统初始化请求结构
type InitRequest struct {
	AdminUsername string `json:"adminUsername"`
	AdminPassword string `json:"adminPassword"`
}

// InitResponse 系统初始化响应结构
type InitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

