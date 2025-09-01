// backend/version.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// 版本配置文件结构
type VersionConfig struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	BuildTime   string   `json:"build_time"`
	GitCommit   string   `json:"git_commit"`
	GoVersion   string   `json:"go_version"`
	GitHubRepo  string   `json:"github_repo"`
	Homepage    string   `json:"homepage"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Changelog   map[string]struct {
		Date    string   `json:"date"`
		Changes []string `json:"changes"`
	} `json:"changelog"`
}

// 全局版本配置
var versionConfig *VersionConfig

// 初始化版本配置
func init() {
	loadVersionConfig()
}

// 加载版本配置
func loadVersionConfig() {
	// 尝试从不同位置加载配置文件
	configPaths := []string{
		"../version.json",
		"./version.json",
		"/opt/clashlink/version.json",
	}

	for _, path := range configPaths {
		if config, err := loadVersionFromFile(path); err == nil {
			versionConfig = config
			return
		}
	}

	// 如果加载失败，使用默认配置
	versionConfig = &VersionConfig{
		Name:       "ClashLink",
		Version:    "1.0.0",
		GitHubRepo: "your-username/clashlink", // 替换为你的GitHub仓库
	}
}

// 从文件加载版本配置
func loadVersionFromFile(path string) (*VersionConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config VersionConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// 获取当前版本
func getCurrentVersion() string {
	if versionConfig != nil {
		return versionConfig.Version
	}
	return "1.0.0"
}

// 获取GitHub仓库
func getGitHubRepo() string {
	if versionConfig != nil && versionConfig.GitHubRepo != "" {
		return versionConfig.GitHubRepo
	}
	return "your-username/clashlink"
}

// VersionInfo 版本信息结构
type VersionInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	UpdateURL      string `json:"update_url"`
	ReleaseNotes   string `json:"release_notes"`
	PublishedAt    string `json:"published_at"`
}

// GitHubRelease GitHub发布信息结构
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
}

// CheckVersionHandler 检查版本更新的API处理器
func CheckVersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只支持GET方法", http.StatusMethodNotAllowed)
		return
	}

	versionInfo, err := getVersionInfo()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VersionInfo{
			CurrentVersion: getCurrentVersion(),
			LatestVersion:  getCurrentVersion(),
			HasUpdate:      false,
			UpdateURL:      "",
			ReleaseNotes:   "",
			PublishedAt:    "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionInfo)
}

// getVersionInfo 获取版本信息
func getVersionInfo() (*VersionInfo, error) {
	// 获取GitHub最新发布版本
	latestRelease, err := getLatestGitHubRelease()
	if err != nil {
		return nil, fmt.Errorf("获取最新版本失败: %v", err)
	}

	// 比较版本
	hasUpdate := compareVersions(getCurrentVersion(), latestRelease.TagName)

	return &VersionInfo{
		CurrentVersion: getCurrentVersion(),
		LatestVersion:  latestRelease.TagName,
		HasUpdate:      hasUpdate,
		UpdateURL:      latestRelease.HTMLURL,
		ReleaseNotes:   latestRelease.Body,
		PublishedAt:    latestRelease.PublishedAt,
	}, nil
}

// getLatestGitHubRelease 获取GitHub最新发布版本
func getLatestGitHubRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", getGitHubRepo())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent头，GitHub API要求
	req.Header.Set("User-Agent", "ClashLink-Updater/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	// 跳过预发布版本和草稿
	if release.Prerelease || release.Draft {
		return nil, fmt.Errorf("最新版本是预发布版本或草稿")
	}

	return &release, nil
}

// compareVersions 比较版本号，返回是否有更新
func compareVersions(current, latest string) bool {
	// 简单的版本比较，去除 'v' 前缀
	currentClean := strings.TrimPrefix(current, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	// 如果版本号不同，则认为有更新
	return currentClean != latestClean && latestClean != ""
}

// GetCurrentVersionHandler 获取当前版本信息
func GetCurrentVersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只支持GET方法", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"version": getCurrentVersion(),
		"name":    "ClashLink",
	}

	if versionConfig != nil {
		response["name"] = versionConfig.Name
		response["build_time"] = versionConfig.BuildTime
		response["git_commit"] = versionConfig.GitCommit
		response["go_version"] = versionConfig.GoVersion
		response["description"] = versionConfig.Description
		response["features"] = versionConfig.Features
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
