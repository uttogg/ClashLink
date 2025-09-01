// backend/parser.go
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ProxyNode 代理节点结构
type ProxyNode struct {
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Server   string            `json:"server"`
	Port     int               `json:"port"`
	UUID     string            `json:"uuid,omitempty"`
	AlterID  int               `json:"alterId,omitempty"`
	Cipher   string            `json:"cipher,omitempty"`
	Network  string            `json:"network,omitempty"`
	TLS      bool              `json:"tls,omitempty"`
	Path     string            `json:"path,omitempty"`
	Host     string            `json:"host,omitempty"`
	SNI      string            `json:"sni,omitempty"`
	Flow     string            `json:"flow,omitempty"`
	Security string            `json:"security,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// VMess配置结构
type VMessConfig struct {
	Version string `json:"v"`
	PS      string `json:"ps"`
	Add     string `json:"add"`
	Port    string `json:"port"`
	ID      string `json:"id"`
	Aid     string `json:"aid"`
	Net     string `json:"net"`
	Type    string `json:"type"`
	Host    string `json:"host"`
	Path    string `json:"path"`
	TLS     string `json:"tls"`
	SNI     string `json:"sni"`
}

// ParseProxyLinks 解析代理链接
func ParseProxyLinks(links string) ([]ProxyNode, error) {
	var nodes []ProxyNode
	lines := strings.Split(strings.TrimSpace(links), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := parseProxyLink(line, i+1)
		if err != nil {
			continue // 跳过无效的链接，而不是返回错误
		}
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到有效的代理链接")
	}

	return nodes, nil
}

// parseProxyLink 解析单个代理链接
func parseProxyLink(link string, index int) (ProxyNode, error) {
	if strings.HasPrefix(link, "vmess://") {
		return parseVMess(link, index)
	} else if strings.HasPrefix(link, "vless://") {
		return parseVLess(link, index)
	}
	return ProxyNode{}, fmt.Errorf("不支持的协议")
}

// parseVMess 解析VMess链接
func parseVMess(link string, index int) (ProxyNode, error) {
	// 移除vmess://前缀
	encoded := strings.TrimPrefix(link, "vmess://")
	
	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// 尝试URL安全的Base64解码
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return ProxyNode{}, fmt.Errorf("VMess链接解码失败")
		}
	}

	// 解析JSON
	var config VMessConfig
	if err := json.Unmarshal(decoded, &config); err != nil {
		return ProxyNode{}, fmt.Errorf("VMess配置解析失败")
	}

	// 转换端口
	port, err := strconv.Atoi(config.Port)
	if err != nil {
		return ProxyNode{}, fmt.Errorf("无效的端口号")
	}

	// 转换AlterId
	alterId := 0
	if config.Aid != "" {
		alterId, _ = strconv.Atoi(config.Aid)
	}

	name := config.PS
	if name == "" {
		name = fmt.Sprintf("VMess-%d", index)
	}

	node := ProxyNode{
		Type:    "vmess",
		Name:    name,
		Server:  config.Add,
		Port:    port,
		UUID:    config.ID,
		AlterID: alterId,
		Cipher:  "auto",
		Network: config.Net,
		TLS:     config.TLS == "tls",
	}

	// 设置传输层参数
	if config.Net == "ws" {
		node.Path = config.Path
		node.Host = config.Host
	}

	if config.SNI != "" {
		node.SNI = config.SNI
	}

	return node, nil
}

// parseVLess 解析VLess链接
func parseVLess(link string, index int) (ProxyNode, error) {
	// 解析URL
	u, err := url.Parse(link)
	if err != nil {
		return ProxyNode{}, fmt.Errorf("VLess链接格式错误")
	}

	// 提取基本信息
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return ProxyNode{}, fmt.Errorf("无效的端口号")
	}

	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("VLess-%d", index)
	}

	query := u.Query()
	
	node := ProxyNode{
		Type:     "vless",
		Name:     name,
		Server:   u.Hostname(),
		Port:     port,
		UUID:     u.User.Username(),
		Network:  query.Get("type"),
		Security: query.Get("security"),
		Flow:     query.Get("flow"),
	}

	// 设置TLS
	if node.Security == "tls" || node.Security == "reality" {
		node.TLS = true
		node.SNI = query.Get("sni")
	}

	// 设置传输层参数
	if node.Network == "ws" {
		node.Path = query.Get("path")
		node.Host = query.Get("host")
	} else if node.Network == "grpc" {
		node.Path = query.Get("serviceName")
	}

	return node, nil
}

