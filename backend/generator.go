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
	Links          string `json:"links"`
	CheckNodes     bool   `json:"checkNodes"`
	OnlyOnline     bool   `json:"onlyOnline"`
	ConfigName     string `json:"configName"`
	// 自定义配置选项
	MixedPort      int    `json:"mixedPort"`
	ControllerPort int    `json:"controllerPort"`
	AllowLan       bool   `json:"allowLan"`
	LogLevel       string `json:"logLevel"`
	DNSMode        string `json:"dnsMode"`
	EnableIPv6     bool   `json:"enableIPv6"`
	CustomRules    string `json:"customRules"`
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

	clashConfig := GenerateClashConfig(finalNodes, configName, req)

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

// SaveConfigRequest 保存配置请求结构
type SaveConfigRequest struct {
	ConfigContent string `json:"configContent"`
	Filename      string `json:"filename"`
}

// SaveConfigHandler 处理保存配置请求
func SaveConfigHandler(w http.ResponseWriter, r *http.Request) {
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

	var req SaveConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证配置内容
	if strings.TrimSpace(req.ConfigContent) == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "配置内容不能为空",
		})
		return
	}

	// 简单的YAML格式验证
	if !strings.Contains(req.ConfigContent, "proxies:") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "配置文件必须包含 proxies 部分",
		})
		return
	}

	// 确定文件名
	filename := req.Filename
	if filename == "" {
		filename = fmt.Sprintf("clash_config_%s_%d.yaml", user.Username, time.Now().Unix())
	}

	// 确保文件名以.yaml结尾
	if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
		filename += ".yaml"
	}

	// 确定保存路径
	subscriptionDir := "../subscriptions"
	if _, err := os.Stat("/app"); err == nil {
		subscriptionDir = "/app/subscriptions"
	}
	
	filePath := filepath.Join(subscriptionDir, filename)

	// 保存文件
	if err := os.WriteFile(filePath, []byte(req.ConfigContent), 0644); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("保存配置文件失败: %v", err),
		})
		return
	}

	// 生成订阅URL
	subscriptionURL := fmt.Sprintf("http://%s/subscriptions/%s", r.Host, filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"message":         "配置保存成功",
		"filename":        filename,
		"subscriptionUrl": subscriptionURL,
	})
}

// GenerateClashConfig 生成Clash配置文件
func GenerateClashConfig(nodes []ProxyNode, configName string, config GenerateRequest) string {
	var configBuilder strings.Builder
	
	// 设置默认值
	mixedPort := config.MixedPort
	if mixedPort == 0 {
		mixedPort = 7890
	}
	
	controllerPort := config.ControllerPort
	if controllerPort == 0 {
		controllerPort = 9090
	}
	
	logLevel := config.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}
	
	dnsMode := config.DNSMode
	if dnsMode == "" {
		dnsMode = "fake-ip"
	}
	
	// 基础配置
	configBuilder.WriteString(fmt.Sprintf(`# Clash配置文件 - %s
# 生成时间: %s
# ClashLink 自动生成

# 基础配置
mixed-port: %d
allow-lan: %t
bind-address: '*'
mode: rule
log-level: %s
external-controller: '127.0.0.1:%d'

# DNS 配置
dns:
  enable: true
  ipv6: %t
  default-nameserver:
    - 223.5.5.5
    - 114.114.114.114
    - 8.8.8.8
  enhanced-mode: %s
  fake-ip-range: 198.18.0.1/16
  use-hosts: true
  nameserver:
    - https://doh.pub/dns-query
    - https://dns.alidns.com/dns-query
    - https://cloudflare-dns.com/dns-query

# 代理节点
proxies:
`, configName, time.Now().Format("2006-01-02 15:04:05"), mixedPort, config.AllowLan, logLevel, controllerPort, config.EnableIPv6, dnsMode))

	// 代理节点配置
	proxyNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		proxyNames = append(proxyNames, node.Name)
		configBuilder.WriteString(generateProxyConfig(node))
	}

	// 代理组配置
	configBuilder.WriteString("\n# 代理组\nproxy-groups:\n")
	configBuilder.WriteString("  - name: \"🚀 节点选择\"\n")
	configBuilder.WriteString("    type: select\n")
	configBuilder.WriteString("    proxies:\n")
	configBuilder.WriteString("      - \"♻️ 自动选择\"\n")
	configBuilder.WriteString("      - \"🎯 全球直连\"\n")
	for _, name := range proxyNames {
		configBuilder.WriteString(fmt.Sprintf("      - \"%s\"\n", name))
	}

	configBuilder.WriteString("  - name: \"♻️ 自动选择\"\n")
	configBuilder.WriteString("    type: url-test\n")
	configBuilder.WriteString("    url: http://www.gstatic.com/generate_204\n")
	configBuilder.WriteString("    interval: 300\n")
	configBuilder.WriteString("    tolerance: 50\n")
	configBuilder.WriteString("    proxies:\n")
	for _, name := range proxyNames {
		configBuilder.WriteString(fmt.Sprintf("      - \"%s\"\n", name))
	}

	configBuilder.WriteString("  - name: \"🎯 全球直连\"\n")
	configBuilder.WriteString("    type: select\n")
	configBuilder.WriteString("    proxies:\n")
	configBuilder.WriteString("      - \"DIRECT\"\n")
	configBuilder.WriteString("      - \"🚀 节点选择\"\n")

	// 规则配置
	configBuilder.WriteString(`
# 分流规则
rules:
  - DOMAIN-SUFFIX,local,DIRECT
  - IP-CIDR,127.0.0.0/8,DIRECT
  - IP-CIDR,172.16.0.0/12,DIRECT
  - IP-CIDR,192.168.0.0/16,DIRECT
  - IP-CIDR,10.0.0.0/8,DIRECT
  - IP-CIDR,17.0.0.0/8,DIRECT
  - IP-CIDR,100.64.0.0/10,DIRECT
  - DOMAIN-SUFFIX,cn,🎯 全球直连
  - GEOIP,CN,🎯 全球直连`)

	// 添加自定义规则
	if config.CustomRules != "" {
		configBuilder.WriteString("\n  # 自定义规则\n")
		// 按行分割自定义规则
		rules := strings.Split(strings.TrimSpace(config.CustomRules), "\n")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule != "" && !strings.HasPrefix(rule, "#") {
				configBuilder.WriteString(fmt.Sprintf("  %s\n", rule))
			}
		}
	}

	configBuilder.WriteString(`
  - MATCH,🚀 节点选择
`)

	return configBuilder.String()
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
		
		if node.Network == "ws" && node.WSOpts != nil {
			if node.WSOpts.Path != "" {
				config.WriteString(fmt.Sprintf("    ws-path: %s\n", node.WSOpts.Path))
			}
			if node.WSOpts.Headers != nil && node.WSOpts.Headers["Host"] != "" {
				config.WriteString(fmt.Sprintf("    ws-headers:\n      Host: %s\n", node.WSOpts.Headers["Host"]))
			}
		} else if node.Network == "grpc" && node.GRPCopts != nil && node.GRPCopts.ServiceName != "" {
			config.WriteString(fmt.Sprintf("    grpc-service-name: %s\n", node.GRPCopts.ServiceName))
		} else if (node.Network == "h2" || node.Network == "http") && node.HTTPOpts != nil {
			if node.HTTPOpts.Path != "" {
				config.WriteString(fmt.Sprintf("    h2-opts:\n      path: %s\n", node.HTTPOpts.Path))
			}
			if node.HTTPOpts.Headers != nil && node.HTTPOpts.Headers["Host"] != "" {
				config.WriteString(fmt.Sprintf("      host: %s\n", node.HTTPOpts.Headers["Host"]))
			}
		}
	}
	
	if node.TLS != nil && *node.TLS {
		config.WriteString("    tls: true\n")
		if node.SNI != "" {
			config.WriteString(fmt.Sprintf("    servername: %s\n", node.SNI))
		}
		if node.SkipCertVerify {
			config.WriteString("    skip-cert-verify: true\n")
		}
	}

	if node.Flow != "" {
		config.WriteString(fmt.Sprintf("    flow: %s\n", node.Flow))
	}

	config.WriteString("    udp: true\n")
	config.WriteString("\n")

	return config.String()
}
