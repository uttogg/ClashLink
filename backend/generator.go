// backend/generator.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateRequest 生成订阅请求结构
type GenerateRequest struct {
	Links      string `json:"links"`
	CheckNodes bool   `json:"checkNodes"`
	OnlyOnline bool   `json:"onlyOnline"`
	ConfigName string `json:"configName"`
}

// GenerateResponse 生成订阅响应结构
type GenerateResponse struct {
	Success         bool           `json:"success"`
	Message         string         `json:"message"`
	SubscriptionURL string         `json:"subscriptionUrl,omitempty"`
	NodeStatuses    []NodeStatus   `json:"nodeStatuses,omitempty"`
	Summary         map[string]int `json:"summary,omitempty"`
	ConfigContent   string         `json:"configContent,omitempty"`
}

// GenerateSubscriptionHandler 处理生成订阅请求
func GenerateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户信息
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "无法获取用户信息", http.StatusUnauthorized)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证输入
	if strings.TrimSpace(req.Links) == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: "请提供代理链接",
		})
		return
	}

	// 解析代理链接
	nodes, err := ParseProxyLinks(req.Links)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: fmt.Sprintf("解析代理链接失败: %v", err),
		})
		return
	}

	response := GenerateResponse{
		Success: true,
		Message: fmt.Sprintf("成功解析 %d 个节点", len(nodes)),
	}

	// 检查节点连通性
	var finalNodes []ProxyNode
	if req.CheckNodes {
		statuses := CheckNodesConnectivity(nodes)
		response.NodeStatuses = statuses
		response.Summary = GetConnectivitySummary(statuses)

		if req.OnlyOnline {
			finalNodes = FilterOnlineNodes(statuses)
			if len(finalNodes) == 0 {
				response.Success = false
				response.Message = "没有在线节点"
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			finalNodes = nodes
		}
	} else {
		finalNodes = nodes
	}

	// 生成Clash配置
	configName := req.ConfigName
	if configName == "" {
		configName = fmt.Sprintf("clash_config_%s_%d", user.Username, time.Now().Unix())
	}

	clashConfig := GenerateClashConfig(finalNodes, configName)

	// 保存配置文件
	filename := fmt.Sprintf("%s.yaml", configName)
	subscriptionDir := "../subscriptions"
	// 在Docker环境中使用绝对路径
	if _, err := os.Stat("/app"); err == nil {
		subscriptionDir = "/app/subscriptions"
	}
	filepath := filepath.Join(subscriptionDir, filename)

	if err := os.WriteFile(filepath, []byte(clashConfig), 0644); err != nil {
		response.Success = false
		response.Message = fmt.Sprintf("保存配置文件失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 生成订阅URL
	subscriptionURL := fmt.Sprintf("http://%s/subscriptions/%s", r.Host, filename)
	response.SubscriptionURL = subscriptionURL
	response.ConfigContent = clashConfig
	response.Message = fmt.Sprintf("成功生成包含 %d 个节点的配置", len(finalNodes))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetSubscriptionHandler 处理重置订阅请求
func ResetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户信息
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "无法获取用户信息", http.StatusUnauthorized)
		return
	}

	// 删除用户的所有订阅文件
	subscriptionDir := "../subscriptions"
	if _, err := os.Stat("/app"); err == nil {
		subscriptionDir = "/app/subscriptions"
	}

	// 查找用户的订阅文件
	files, err := os.ReadDir(subscriptionDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "读取订阅目录失败",
		})
		return
	}

	deletedCount := 0
	userPrefix := fmt.Sprintf("clash_config_%s_", user.Username)

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), userPrefix) && strings.HasSuffix(file.Name(), ".yaml") {
			filePath := filepath.Join(subscriptionDir, file.Name())
			if err := os.Remove(filePath); err == nil {
				deletedCount++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("已删除 %d 个订阅文件", deletedCount),
		"deleted_count": deletedCount,
	})
}

// GenerateClashConfig 生成Clash配置文件
func GenerateClashConfig(nodes []ProxyNode, configName string) string {
	var config strings.Builder

	// 基础配置
	config.WriteString(fmt.Sprintf(`# Clash配置文件 - %s
# 生成时间: %s

mixed-port: 7890
allow-lan: true
bind-address: '*'
mode: rule
log-level: info
external-controller: '127.0.0.1:9090'

dns:
  enable: true
  ipv6: false
  default-nameserver:
    - 223.5.5.5
    - 114.114.114.114
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  use-hosts: true
  nameserver:
    - https://doh.pub/dns-query
    - https://dns.alidns.com/dns-query

proxies:
`, configName, time.Now().Format("2006-01-02 15:04:05")))

	// 代理节点配置
	proxyNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		proxyNames = append(proxyNames, node.Name)
		config.WriteString(generateProxyConfig(node))
	}

	// 代理组配置
	config.WriteString("\nproxy-groups:\n")
	config.WriteString("  - name: \"🚀 节点选择\"\n")
	config.WriteString("    type: select\n")
	config.WriteString("    proxies:\n")
	config.WriteString("      - \"♻️ 自动选择\"\n")
	config.WriteString("      - \"DIRECT\"\n")
	for _, name := range proxyNames {
		config.WriteString(fmt.Sprintf("      - \"%s\"\n", name))
	}

	config.WriteString("  - name: \"♻️ 自动选择\"\n")
	config.WriteString("    type: url-test\n")
	config.WriteString("    url: http://www.gstatic.com/generate_204\n")
	config.WriteString("    interval: 300\n")
	config.WriteString("    proxies:\n")
	for _, name := range proxyNames {
		config.WriteString(fmt.Sprintf("      - \"%s\"\n", name))
	}

	config.WriteString("  - name: \"🎯 全球直连\"\n")
	config.WriteString("    type: select\n")
	config.WriteString("    proxies:\n")
	config.WriteString("      - \"DIRECT\"\n")
	config.WriteString("      - \"🚀 节点选择\"\n")

	// 规则配置
	config.WriteString(`
rules:
  - DOMAIN-SUFFIX,local,DIRECT
  - IP-CIDR,127.0.0.0/8,DIRECT
  - IP-CIDR,172.16.0.0/12,DIRECT
  - IP-CIDR,192.168.0.0/16,DIRECT
  - IP-CIDR,10.0.0.0/8,DIRECT
  - IP-CIDR,17.0.0.0/8,DIRECT
  - IP-CIDR,100.64.0.0/10,DIRECT
  - DOMAIN-SUFFIX,cn,🎯 全球直连
  - GEOIP,CN,🎯 全球直连
  - MATCH,🚀 节点选择
`)

	return config.String()
}

// generateProxyConfig 生成单个代理配置
func generateProxyConfig(node ProxyNode) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("  - name: \"%s\"\n", node.Name))
	config.WriteString(fmt.Sprintf("    type: %s\n", node.Type))
	config.WriteString(fmt.Sprintf("    server: %s\n", node.Server))
	config.WriteString(fmt.Sprintf("    port: %d\n", node.Port))

	if node.UUID != "" {
		config.WriteString(fmt.Sprintf("    uuid: %s\n", node.UUID))
	}

	if node.Type == "vmess" {
		config.WriteString(fmt.Sprintf("    alterId: %d\n", node.AlterID))
		config.WriteString(fmt.Sprintf("    cipher: %s\n", node.Cipher))
	}

	if node.Network != "" && node.Network != "tcp" {
		config.WriteString(fmt.Sprintf("    network: %s\n", node.Network))

		if node.Network == "ws" {
			if node.Path != "" {
				config.WriteString(fmt.Sprintf("    ws-path: %s\n", node.Path))
			}
			if node.Host != "" {
				config.WriteString(fmt.Sprintf("    ws-headers:\n      Host: %s\n", node.Host))
			}
		} else if node.Network == "grpc" && node.Path != "" {
			config.WriteString(fmt.Sprintf("    grpc-service-name: %s\n", node.Path))
		}
	}

	if node.TLS {
		config.WriteString("    tls: true\n")
		if node.SNI != "" {
			config.WriteString(fmt.Sprintf("    servername: %s\n", node.SNI))
		}
	}

	if node.Flow != "" {
		config.WriteString(fmt.Sprintf("    flow: %s\n", node.Flow))
	}

	config.WriteString("    udp: true\n")
	config.WriteString("\n")

	return config.String()
}
