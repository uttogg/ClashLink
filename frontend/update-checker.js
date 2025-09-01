// frontend/update-checker.js
class UpdateChecker {
    constructor() {
        this.checkInterval = 30 * 60 * 1000; // 30分钟检查一次
        this.lastCheckTime = 0;
        this.updateNotification = null;
        this.isChecking = false;
        
        this.init();
    }

    init() {
        // 页面加载后立即检查一次
        setTimeout(() => this.checkForUpdates(), 2000);
        
        // 设置定期检查
        setInterval(() => this.checkForUpdates(), this.checkInterval);
        
        // 创建更新通知元素
        this.createUpdateNotification();
    }

    // 创建更新通知UI
    createUpdateNotification() {
        const notification = document.createElement('div');
        notification.id = 'update-notification';
        notification.className = 'update-notification';
        notification.style.display = 'none';
        
        notification.innerHTML = `
            <div class="update-content">
                <div class="update-icon">🚀</div>
                <div class="update-text">
                    <div class="update-title">发现新版本！</div>
                    <div class="update-version">v<span id="latest-version"></span> 现已可用</div>
                </div>
                <div class="update-actions">
                    <button id="view-update" class="update-btn primary">查看更新</button>
                    <button id="dismiss-update" class="update-btn secondary">稍后提醒</button>
                    <button id="close-update" class="update-btn close">&times;</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(notification);
        this.updateNotification = notification;
        
        // 绑定事件
        this.bindUpdateEvents();
    }

    // 绑定更新通知事件
    bindUpdateEvents() {
        const viewBtn = document.getElementById('view-update');
        const dismissBtn = document.getElementById('dismiss-update');
        const closeBtn = document.getElementById('close-update');
        
        if (viewBtn) {
            viewBtn.addEventListener('click', () => this.viewUpdate());
        }
        
        if (dismissBtn) {
            dismissBtn.addEventListener('click', () => this.dismissUpdate());
        }
        
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.hideNotification());
        }
    }

    // 检查更新
    async checkForUpdates() {
        if (this.isChecking) return;
        
        this.isChecking = true;
        const now = Date.now();
        
        try {
            const response = await fetch('/api/check-update', {
                method: 'GET',
                headers: {
                    'Cache-Control': 'no-cache'
                }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const versionInfo = await response.json();
            this.lastCheckTime = now;
            
            if (versionInfo.has_update) {
                this.showUpdateNotification(versionInfo);
            }
            
        } catch (error) {
            console.warn('检查更新失败:', error);
        } finally {
            this.isChecking = false;
        }
    }

    // 显示更新通知
    showUpdateNotification(versionInfo) {
        // 检查是否已经忽略了这个版本
        const dismissedVersion = localStorage.getItem('dismissed_update_version');
        if (dismissedVersion === versionInfo.latest_version) {
            return;
        }
        
        // 更新版本信息
        const versionSpan = document.getElementById('latest-version');
        if (versionSpan) {
            versionSpan.textContent = versionInfo.latest_version;
        }
        
        // 存储更新信息
        this.currentUpdateInfo = versionInfo;
        
        // 显示通知
        if (this.updateNotification) {
            this.updateNotification.style.display = 'block';
            this.updateNotification.classList.add('show');
            
            // 添加动画
            setTimeout(() => {
                this.updateNotification.classList.add('slide-in');
            }, 100);
        }
    }

    // 隐藏通知
    hideNotification() {
        if (this.updateNotification) {
            this.updateNotification.classList.remove('slide-in');
            setTimeout(() => {
                this.updateNotification.style.display = 'none';
                this.updateNotification.classList.remove('show');
            }, 300);
        }
    }

    // 查看更新详情
    viewUpdate() {
        if (this.currentUpdateInfo && this.currentUpdateInfo.update_url) {
            // 显示更新详情弹窗
            this.showUpdateModal(this.currentUpdateInfo);
        }
    }

    // 显示更新详情弹窗
    showUpdateModal(updateInfo) {
        // 创建模态框
        const modal = document.createElement('div');
        modal.className = 'update-modal';
        modal.innerHTML = `
            <div class="update-modal-content">
                <div class="update-modal-header">
                    <h2>🎉 新版本可用</h2>
                    <button class="update-modal-close">&times;</button>
                </div>
                <div class="update-modal-body">
                    <div class="version-info">
                        <div class="version-current">
                            <span class="version-label">当前版本:</span>
                            <span class="version-number">${updateInfo.current_version}</span>
                        </div>
                        <div class="version-arrow">→</div>
                        <div class="version-latest">
                            <span class="version-label">最新版本:</span>
                            <span class="version-number">${updateInfo.latest_version}</span>
                        </div>
                    </div>
                    
                    <div class="release-notes">
                        <h3>更新内容:</h3>
                        <div class="release-content">
                            ${this.formatReleaseNotes(updateInfo.release_notes)}
                        </div>
                    </div>
                    
                    <div class="release-date">
                        <small>发布时间: ${this.formatDate(updateInfo.published_at)}</small>
                    </div>
                </div>
                <div class="update-modal-footer">
                    <button class="update-btn secondary" onclick="this.closest('.update-modal').remove()">稍后更新</button>
                    <a href="${updateInfo.update_url}" target="_blank" class="update-btn primary">立即更新</a>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // 绑定关闭事件
        const closeBtn = modal.querySelector('.update-modal-close');
        closeBtn.addEventListener('click', () => modal.remove());
        
        // 点击背景关闭
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
        // 隐藏底部通知
        this.hideNotification();
    }

    // 格式化发布说明
    formatReleaseNotes(notes) {
        if (!notes) return '<p>暂无更新说明</p>';
        
        // 简单的 Markdown 转 HTML
        return notes
            .replace(/### (.*)/g, '<h4>$1</h4>')
            .replace(/## (.*)/g, '<h3>$1</h3>')
            .replace(/# (.*)/g, '<h2>$1</h2>')
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/- (.*)/g, '<li>$1</li>')
            .replace(/(<li>.*<\/li>)/gs, '<ul>$1</ul>')
            .replace(/\n/g, '<br>');
    }

    // 格式化日期
    formatDate(dateString) {
        if (!dateString) return '未知';
        
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // 忽略此版本更新
    dismissUpdate() {
        if (this.currentUpdateInfo) {
            localStorage.setItem('dismissed_update_version', this.currentUpdateInfo.latest_version);
            this.hideNotification();
        }
    }

    // 手动检查更新
    async manualCheck() {
        if (this.isChecking) {
            return { checking: true };
        }
        
        await this.checkForUpdates();
        return { 
            checking: false, 
            lastCheck: this.lastCheckTime,
            hasUpdate: !!this.currentUpdateInfo 
        };
    }
}

// 全局更新检查器实例
let updateChecker;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    updateChecker = new UpdateChecker();
    
    // 将检查器添加到全局作用域，供其他脚本使用
    window.updateChecker = updateChecker;
});

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UpdateChecker;
}
