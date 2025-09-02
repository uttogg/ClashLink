// backend/middleware.go
package main

import (
	"context"
	"net/http"
	"strings"
)

// JWTMiddleware JWT认证中间件
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 对于页面访问（非API），重定向到登录页面而不是返回错误
		isPageRequest := r.Header.Get("Accept") != "" &&
			(strings.Contains(r.Header.Get("Accept"), "text/html") ||
				strings.Contains(r.Header.Get("Accept"), "*/*"))

		// 从请求头获取Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			if isPageRequest {
				// 页面请求重定向到登录页
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			} else {
				// API请求返回错误
				http.Error(w, "缺少认证令牌", http.StatusUnauthorized)
				return
			}
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			http.Error(w, "无效的认证格式", http.StatusUnauthorized)
			return
		}

		// 提取令牌
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			http.Error(w, "令牌为空", http.StatusUnauthorized)
			return
		}

		// 验证令牌
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "无效的令牌", http.StatusUnauthorized)
			return
		}

		// 将用户信息添加到请求上下文中
		ctx := context.WithValue(r.Context(), "user", claims)
		r = r.WithContext(ctx)

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext 从请求上下文中获取用户信息
func GetUserFromContext(r *http.Request) (*Claims, bool) {
	user, ok := r.Context().Value("user").(*Claims)
	return user, ok
}
