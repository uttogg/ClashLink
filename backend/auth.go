// backend/auth.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT密钥 - 在生产环境中应该从环境变量读取
var jwtSecret = []byte("your-secret-key-change-this-in-production")

// Claims JWT声明结构
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// InitSystemHandler 处理系统初始化
func InitSystemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 检查系统是否已经初始化
	initialized, err := IsSystemInitialized()
	if err != nil {
		http.Error(w, "检查系统状态失败", http.StatusInternalServerError)
		return
	}

	if initialized {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(InitResponse{
			Success: false,
			Message: "系统已经初始化",
		})
		return
	}

	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.AdminUsername == "" || req.AdminPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(InitResponse{
			Success: false,
			Message: "管理员用户名和密码不能为空",
		})
		return
	}

	if len(req.AdminUsername) < 3 || len(req.AdminPassword) < 6 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(InitResponse{
			Success: false,
			Message: "用户名至少3位，密码至少6位",
		})
		return
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "密码处理失败", http.StatusInternalServerError)
		return
	}

	// 创建管理员用户
	if err := CreateUser(req.AdminUsername, string(hashedPassword), true); err != nil {
		http.Error(w, "创建管理员失败", http.StatusInternalServerError)
		return
	}

	// 设置系统为已初始化
	if err := SetSystemInitialized(); err != nil {
		http.Error(w, "设置初始化状态失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(InitResponse{
		Success: true,
		Message: "系统初始化成功",
	})
}

// RegisterHandler 处理用户注册
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 检查系统是否已经初始化
	initialized, err := IsSystemInitialized()
	if err != nil {
		http.Error(w, "检查系统状态失败", http.StatusInternalServerError)
		return
	}

	if !initialized {
		http.Error(w, "系统未初始化", http.StatusForbidden)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.Username == "" || req.Password == "" {
		http.Error(w, "用户名和密码不能为空", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		http.Error(w, "用户名至少3位，密码至少6位", http.StatusBadRequest)
		return
	}

	// 检查用户是否已存在
	if _, err := GetUserByUsername(req.Username); err == nil {
		http.Error(w, "用户名已存在", http.StatusConflict)
		return
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "密码处理失败", http.StatusInternalServerError)
		return
	}

	// 创建用户
	if err := CreateUser(req.Username, string(hashedPassword), false); err != nil {
		http.Error(w, "创建用户失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "注册成功",
	})
}

// LoginHandler 处理用户登录
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 检查系统是否已经初始化
	initialized, err := IsSystemInitialized()
	if err != nil {
		http.Error(w, "检查系统状态失败", http.StatusInternalServerError)
		return
	}

	if !initialized {
		http.Error(w, "系统未初始化", http.StatusForbidden)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证输入
	if req.Username == "" || req.Password == "" {
		http.Error(w, "用户名和密码不能为空", http.StatusBadRequest)
		return
	}

	// 查找用户
	user, err := GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}

	// 生成JWT令牌
	token, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "生成令牌失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:   token,
		Message: "登录成功",
	})
}

// GenerateJWT 生成JWT令牌
func GenerateJWT(userID int, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT 验证JWT令牌
func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

