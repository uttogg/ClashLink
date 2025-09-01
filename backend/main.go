// backend/main.go
package main

import (
	"log"
	"net/http"
	"os"
)

// RootHandler 根路由处理，根据系统状态重定向
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// 检查系统是否已经初始化
	initialized, err := IsSystemInitialized()
	if err != nil {
		http.Error(w, "检查系统状态失败", http.StatusInternalServerError)
		return
	}

	if !initialized {
		// 系统未初始化，重定向到初始化页面
		http.Redirect(w, r, "/init", http.StatusTemporaryRedirect)
		return
	}

	// 系统已初始化，重定向到登录页面
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func main() {
	// 初始化数据库
	if err := InitDB(); err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	// 创建订阅目录
	if err := os.MkdirAll("../subscriptions", 0755); err != nil {
		log.Fatal("创建订阅目录失败:", err)
	}

	// 设置路由
	mux := http.NewServeMux()

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../frontend/"))))
	mux.Handle("/subscriptions/", http.StripPrefix("/subscriptions/", http.FileServer(http.Dir("../subscriptions/"))))

	// 公开路由（无需认证）
	mux.HandleFunc("/", RootHandler)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/login.html")
	})
	mux.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/init.html")
	})
	mux.HandleFunc("/api/init", InitSystemHandler)
	mux.HandleFunc("/api/register", RegisterHandler)
	mux.HandleFunc("/api/login", LoginHandler)
	mux.HandleFunc("/api/version", GetCurrentVersionHandler)
	mux.HandleFunc("/api/check-update", CheckVersionHandler)

	// 保护路由（需要JWT认证）
	mux.Handle("/app", JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/index.html")
	})))
	mux.Handle("/api/generate", JWTMiddleware(http.HandlerFunc(GenerateSubscriptionHandler)))

	log.Println("服务器启动在端口 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
