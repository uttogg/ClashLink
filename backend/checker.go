// backend/checker.go
package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NodeStatus 节点状态
type NodeStatus struct {
	Node     ProxyNode `json:"node"`
	Status   string    `json:"status"`   // "online", "offline", "timeout"
	Latency  int       `json:"latency"`  // 延迟毫秒
	Error    string    `json:"error,omitempty"`
}

// CheckNodesConnectivity 检查节点连通性
func CheckNodesConnectivity(nodes []ProxyNode) []NodeStatus {
	var wg sync.WaitGroup
	results := make([]NodeStatus, len(nodes))
	
	// 并发检查每个节点
	for i, node := range nodes {
		wg.Add(1)
		go func(index int, n ProxyNode) {
			defer wg.Done()
			results[index] = checkSingleNode(n)
		}(i, node)
	}
	
	wg.Wait()
	return results
}

// checkSingleNode 检查单个节点
func checkSingleNode(node ProxyNode) NodeStatus {
	status := NodeStatus{
		Node:    node,
		Status:  "offline",
		Latency: -1,
	}

	// 构建地址
	address := fmt.Sprintf("%s:%d", node.Server, node.Port)
	
	// 设置超时时间
	timeout := 5 * time.Second
	start := time.Now()
	
	// 尝试TCP连接
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		status.Error = err.Error()
		// 检查是否是超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			status.Status = "timeout"
		}
		return status
	}
	defer conn.Close()
	
	// 计算延迟
	latency := time.Since(start)
	status.Status = "online"
	status.Latency = int(latency.Milliseconds())
	
	return status
}

// FilterOnlineNodes 过滤在线节点
func FilterOnlineNodes(statuses []NodeStatus) []ProxyNode {
	var onlineNodes []ProxyNode
	
	for _, status := range statuses {
		if status.Status == "online" {
			onlineNodes = append(onlineNodes, status.Node)
		}
	}
	
	return onlineNodes
}

// GetConnectivitySummary 获取连通性摘要
func GetConnectivitySummary(statuses []NodeStatus) map[string]int {
	summary := map[string]int{
		"total":   len(statuses),
		"online":  0,
		"offline": 0,
		"timeout": 0,
	}
	
	for _, status := range statuses {
		summary[status.Status]++
	}
	
	return summary
}

