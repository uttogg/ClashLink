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

// ProxyNode 结构体用于存储解析后的节点信息，以便转换为 Clash YAML 格式
type ProxyNode struct {
	Name           string    `yaml:"name"`
	Type           string    `yaml:"type"` // vmess, vless, ss, trojan, etc.
	Server         string    `yaml:"server"`
	Port           int       `yaml:"port"`
	UUID           string    `yaml:"uuid,omitempty"`     // For vmess/vless
	Password       string    `yaml:"password,omitempty"` // For ss/trojan/vless (VLESS password is UUID)
	AlterID        int       `yaml:"alterId,omitempty"`  // For vmess
	Cipher         string    `yaml:"cipher,omitempty"`   // For vmess/ss
	TLS            *bool     `yaml:"tls,omitempty"`      // Pointer to bool to differentiate false from omitted
	SkipCertVerify bool      `yaml:"skip-cert-verify,omitempty"`
	Network        string    `yaml:"network,omitempty"` // tcp, ws, http, h2, grpc
	HTTPOpts       *struct { // For http network (e.g. VLESS h2)
		Method  string            `yaml:"method,omitempty"`
		Headers map[string]string `yaml:"headers,omitempty"`
		Path    string            `yaml:"path,omitempty"`
	} `yaml:"http-opts,omitempty"`
	WSOpts *struct { // For ws network
		Path    string            `yaml:"path"`
		Headers map[string]string `yaml:"headers,omitempty"`
	} `yaml:"ws-opts,omitempty"`
	GRPCopts *struct { // For grpc network
		ServiceName string `yaml:"service-name,omitempty"`
		Mode        string `yaml:"mode,omitempty"`
	} `yaml:"grpc-opts,omitempty"`
	Flow        string `yaml:"flow,omitempty"`               // For VLESS XTLS/Reality
	UDP         bool   `yaml:"udp,omitempty"`                // For UDP forwarding
	SNI         string `yaml:"servername,omitempty"`         // TLS SNI
	Fingerprint string `yaml:"client-fingerprint,omitempty"` // Reality fingerprint
}

// VMessLinkRaw 结构体用于解析 VMess 链接中的 JSON 内容
type VMessLinkRaw struct {
	V    string `json:"v"`
	Ps   string `json:"ps"`   // 备注/名称
	Add  string `json:"add"`  // 服务器地址
	Port string `json:"port"` // 端口
	ID   string `json:"id"`   // UUID
	Aid  string `json:"aid"`  // AlterId
	Scy  string `json:"scy"`  // 加密方式
	Net  string `json:"net"`  // 网络类型 (tcp, ws, h2, ...)
	Type string `json:"type"` // 加密类型 (none, http, ...)
	Host string `json:"host"` // Host header (for ws/h2)
	Path string `json:"path"` // Path (for ws/h2)
	TLS  string `json:"tls"`  // TLS 类型 (tls, none, ...)
	SNI  string `json:"sni"`  // TLS SNI
}

// parseVMessLink 解析单个 VMess 链接并转换为 ProxyNode 结构体
func parseVMessLink(rawVMessLink string) (*ProxyNode, error) {
	if !strings.HasPrefix(rawVMessLink, "vmess://") {
		return nil, fmt.Errorf("无效的VMess链接格式: %s", rawVMessLink)
	}

	encodedPart := strings.TrimPrefix(rawVMessLink, "vmess://")

	// Base64 解码，尝试多种解码方式
	var decodedBytes []byte
	var err error

	// 尝试标准 Base64
	decodedBytes, err = base64.StdEncoding.DecodeString(encodedPart)
	if err != nil {
		// 尝试 URL 安全的 Base64
		decodedBytes, err = base64.URLEncoding.DecodeString(encodedPart)
		if err != nil {
			// 尝试无填充的 Base64
			decodedBytes, err = base64.RawStdEncoding.DecodeString(encodedPart)
			if err != nil {
				return nil, fmt.Errorf("VMess链接Base64解码失败: %w", err)
			}
		}
	}

	var rawConfig VMessLinkRaw
	if err := json.Unmarshal(decodedBytes, &rawConfig); err != nil {
		return nil, fmt.Errorf("VMess配置JSON解析失败: %w", err)
	}

	// 验证必要字段
	if rawConfig.Add == "" || rawConfig.Port == "" || rawConfig.ID == "" {
		return nil, fmt.Errorf("VMess链接缺少必要字段")
	}

	// 端口转换为 int
	port, err := strconv.Atoi(rawConfig.Port)
	if err != nil {
		return nil, fmt.Errorf("VMess链接端口无效: %s", rawConfig.Port)
	}

	// alterId 转换为 int
	alterID := 0
	if rawConfig.Aid != "" {
		alterID, _ = strconv.Atoi(rawConfig.Aid)
	}

	// 设置默认名称
	name := rawConfig.Ps
	if name == "" {
		name = fmt.Sprintf("VMess-%s:%d", rawConfig.Add, port)
	}

	// TLS 配置
	tlsEnabled := rawConfig.TLS == "tls"

	// 创建节点
	node := &ProxyNode{
		Name:    name,
		Type:    "vmess",
		Server:  rawConfig.Add,
		Port:    port,
		UUID:    rawConfig.ID,
		AlterID: alterID,
		Cipher:  rawConfig.Scy,
		Network: rawConfig.Net,
		TLS:     &tlsEnabled,
		UDP:     true,
	}

	// 设置默认加密方式
	if node.Cipher == "" {
		node.Cipher = "auto"
	}

	// 设置默认网络类型
	if node.Network == "" {
		node.Network = "tcp"
	}

	// 设置 SNI
	if rawConfig.SNI != "" {
		node.SNI = rawConfig.SNI
	}

	// 配置传输层选项
	hostHeader := rawConfig.Host
	if hostHeader == "" && (rawConfig.Net == "ws" || rawConfig.Net == "h2" || rawConfig.Net == "http") {
		hostHeader = rawConfig.Add
	}

	switch rawConfig.Net {
	case "ws":
		node.WSOpts = &struct {
			Path    string            `yaml:"path"`
			Headers map[string]string `yaml:"headers,omitempty"`
		}{
			Path: rawConfig.Path,
			Headers: map[string]string{
				"Host": hostHeader,
			},
		}
	case "h2", "http":
		node.HTTPOpts = &struct {
			Method  string            `yaml:"method,omitempty"`
			Headers map[string]string `yaml:"headers,omitempty"`
			Path    string            `yaml:"path,omitempty"`
		}{
			Method: "GET",
			Path:   rawConfig.Path,
			Headers: map[string]string{
				"Host": hostHeader,
			},
		}
	}

	return node, nil
}

// parseVLESSLink 解析单个 VLESS 链接并转换为 ProxyNode 结构体
func parseVLESSLink(rawVLESSLink string) (*ProxyNode, error) {
	if !strings.HasPrefix(rawVLESSLink, "vless://") {
		return nil, fmt.Errorf("无效的VLESS链接格式: %s", rawVLESSLink)
	}

	// 解析 VLESS URL
	parsedURL, err := url.Parse(rawVLESSLink)
	if err != nil {
		return nil, fmt.Errorf("VLESS链接URL解析失败: %w", err)
	}

	// 提取基本信息
	uuid := parsedURL.User.Username()
	if uuid == "" {
		return nil, fmt.Errorf("VLESS链接缺少UUID")
	}

	server := parsedURL.Hostname()
	if server == "" {
		return nil, fmt.Errorf("VLESS链接缺少服务器地址")
	}

	portStr := parsedURL.Port()
	if portStr == "" {
		return nil, fmt.Errorf("VLESS链接缺少端口")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("VLESS链接端口无效: %s", portStr)
	}

	// 设置节点名称
	name := parsedURL.Fragment
	if name == "" {
		name = fmt.Sprintf("VLESS-%s:%d", server, port)
	}

	// 解析查询参数
	query := parsedURL.Query()

	// 创建基础节点
	node := &ProxyNode{
		Name:     name,
		Type:     "vless",
		Server:   server,
		Port:     port,
		UUID:     uuid,
		Password: uuid, // VLESS 在 Clash 中通常使用 UUID 作为密码
		UDP:      true,
	}

	// 安全配置
	security := query.Get("security")
	tlsEnabled := security == "tls" || security == "reality"
	node.TLS = &tlsEnabled

	// 跳过证书验证
	if query.Get("allowInsecure") == "1" {
		node.SkipCertVerify = true
	}

	// SNI 配置
	if sni := query.Get("sni"); sni != "" {
		node.SNI = sni
	}

	// 指纹配置 (Reality)
	if fp := query.Get("fp"); fp != "" {
		node.Fingerprint = fp
	}

	// 流控配置
	if flow := query.Get("flow"); flow != "" {
		node.Flow = flow
	}

	// 网络传输配置
	network := query.Get("type")
	if network == "" {
		network = "tcp"
	}
	node.Network = network

	// 配置传输层选项
	switch network {
	case "ws":
		wsPath := query.Get("path")
		wsHost := query.Get("host")
		if wsHost == "" {
			wsHost = server
		}

		node.WSOpts = &struct {
			Path    string            `yaml:"path"`
			Headers map[string]string `yaml:"headers,omitempty"`
		}{
			Path: wsPath,
			Headers: map[string]string{
				"Host": wsHost,
			},
		}

	case "grpc":
		serviceName := query.Get("serviceName")
		mode := query.Get("mode")
		if mode == "" {
			mode = "gun" // 默认模式
		}

		node.GRPCopts = &struct {
			ServiceName string `yaml:"service-name,omitempty"`
			Mode        string `yaml:"mode,omitempty"`
		}{
			ServiceName: serviceName,
			Mode:        mode,
		}

	case "h2", "http":
		httpPath := query.Get("path")
		httpHost := query.Get("host")
		if httpHost == "" {
			httpHost = server
		}

		node.HTTPOpts = &struct {
			Method  string            `yaml:"method,omitempty"`
			Headers map[string]string `yaml:"headers,omitempty"`
			Path    string            `yaml:"path,omitempty"`
		}{
			Method: "GET",
			Path:   httpPath,
			Headers: map[string]string{
				"Host": httpHost,
			},
		}
	}

	return node, nil
}

// ParseProxyLinks 解析代理链接字符串，返回 ProxyNode 列表
func ParseProxyLinks(rawLinks string) ([]ProxyNode, error) {
	var parsedNodes []ProxyNode
	var errors []string

	lines := strings.Split(strings.TrimSpace(rawLinks), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var node *ProxyNode
		var err error

		if strings.HasPrefix(line, "vmess://") {
			node, err = parseVMessLink(line)
		} else if strings.HasPrefix(line, "vless://") {
			node, err = parseVLESSLink(line)
		} else {
			errors = append(errors, fmt.Sprintf("第%d行: 未知协议或无效链接", i+1))
			continue
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d行解析失败: %v", i+1, err))
			continue
		}

		parsedNodes = append(parsedNodes, *node)
	}

	// 如果没有成功解析任何节点，返回错误
	if len(parsedNodes) == 0 {
		if len(errors) > 0 {
			return nil, fmt.Errorf("解析失败: %s", strings.Join(errors, "; "))
		}
		return nil, fmt.Errorf("未找到有效的代理链接")
	}

	return parsedNodes, nil
}
