// frontend/script.js
document.addEventListener('DOMContentLoaded', function() {
    // 检查登录状态
    checkAuthStatus();
    
    // 初始化页面
    initializePage();
    
    // 绑定事件
    setupEventListeners();
});

// 检查认证状态
function checkAuthStatus() {
    console.log('检查认证状态...');
    const token = localStorage.getItem('jwt_token');
    
    if (!token) {
        console.log('未找到JWT令牌，重定向到登录页');
        localStorage.removeItem('jwt_token');
        window.location.href = '/login';
        return;
    }
    
    if (!isTokenValid(token)) {
        console.log('JWT令牌无效或已过期，重定向到登录页');
        localStorage.removeItem('jwt_token');
        window.location.href = '/login';
        return;
    }
    
    console.log('JWT令牌有效，显示用户信息');
    // 显示用户信息
    displayUserInfo(token);
}

// 简单的token有效性检查
function isTokenValid(token) {
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const now = Math.floor(Date.now() / 1000);
        return payload.exp > now;
    } catch (e) {
        return false;
    }
}

// 显示用户信息
function displayUserInfo(token) {
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        document.getElementById('username').textContent = payload.username || '用户';
    } catch (e) {
        document.getElementById('username').textContent = '用户';
    }
}

// 初始化页面
function initializePage() {
    // 更新链接计数
    updateLinkCount();
}

// 设置事件监听器
function setupEventListeners() {
    // 退出登录
    document.getElementById('logoutBtn').addEventListener('click', logout);
    
    // 文本框输入监听
    document.getElementById('nodeLinks').addEventListener('input', updateLinkCount);
    
    // 生成订阅按钮
    document.getElementById('generateBtn').addEventListener('click', generateSubscription);
    
    // 复制URL按钮
    document.getElementById('copyUrlBtn').addEventListener('click', copySubscriptionUrl);
    
    // 配置预览切换
    document.getElementById('toggleConfig').addEventListener('click', toggleConfigPreview);
    
    // 下载配置
    document.getElementById('downloadConfig').addEventListener('click', downloadConfig);
    
    // 重置订阅
    document.getElementById('resetSubscription').addEventListener('click', resetSubscription);
    
    // 复选框联动
    const checkNodesCheckbox = document.getElementById('checkNodes');
    const onlyOnlineCheckbox = document.getElementById('onlyOnline');
    
    checkNodesCheckbox.addEventListener('change', function() {
        if (!this.checked) {
            onlyOnlineCheckbox.checked = false;
            onlyOnlineCheckbox.disabled = true;
        } else {
            onlyOnlineCheckbox.disabled = false;
        }
    });
}

// 更新链接计数
function updateLinkCount() {
    const textarea = document.getElementById('nodeLinks');
    const lines = textarea.value.split('\n').filter(line => line.trim().length > 0);
    const count = lines.length;
    document.getElementById('linkCount').textContent = `${count} 个链接`;
}

// 退出登录
function logout() {
    localStorage.removeItem('jwt_token');
    showMessage('已退出登录', 'info');
    setTimeout(() => {
        window.location.href = '/login';
    }, 1000);
}

// 生成订阅
async function generateSubscription() {
    const nodeLinks = document.getElementById('nodeLinks').value.trim();
    if (!nodeLinks) {
        showMessage('请输入节点链接', 'error');
        return;
    }
    
    const generateBtn = document.getElementById('generateBtn');
    const checkNodes = document.getElementById('checkNodes').checked;
    const onlyOnline = document.getElementById('onlyOnline').checked;
    const configName = document.getElementById('configName').value.trim();
    
    // 显示加载状态
    generateBtn.classList.add('loading');
    generateBtn.disabled = true;
    
    try {
        const token = localStorage.getItem('jwt_token');
        const response = await fetch('/api/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                links: nodeLinks,
                checkNodes: checkNodes,
                onlyOnline: onlyOnline,
                configName: configName
            })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            displayResults(data);
            showMessage('订阅生成成功！', 'success');
        } else {
            showMessage(data.message || '生成失败', 'error');
        }
    } catch (error) {
        console.error('生成订阅错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        // 恢复按钮状态
        generateBtn.classList.remove('loading');
        generateBtn.disabled = false;
    }
}

// 显示结果
function displayResults(data) {
    const resultSection = document.getElementById('resultSection');
    const subscriptionUrl = document.getElementById('subscriptionUrl');
    const nodeStatus = document.getElementById('nodeStatus');
    const configContent = document.getElementById('configContent');
    
    // 显示结果区域
    resultSection.style.display = 'block';
    resultSection.scrollIntoView({ behavior: 'smooth' });
    
    // 设置订阅链接
    subscriptionUrl.value = data.subscriptionURL || '';
    
    // 显示节点状态
    if (data.nodeStatuses && data.nodeStatuses.length > 0) {
        displayNodeStatus(data.nodeStatuses, data.summary);
        nodeStatus.style.display = 'block';
    } else {
        nodeStatus.style.display = 'none';
    }
    
    // 设置配置内容
    configContent.textContent = data.configContent || '';
    
    // 存储配置内容供下载使用
    window.currentConfig = {
        content: data.configContent,
        filename: extractFilenameFromUrl(data.subscriptionURL)
    };
}

// 显示节点状态
function displayNodeStatus(statuses, summary) {
    const statusSummary = document.getElementById('statusSummary');
    const statusList = document.getElementById('statusList');
    
    // 显示摘要
    if (summary) {
        statusSummary.innerHTML = `
            <div class="summary-item total">总计: ${summary.total}</div>
            <div class="summary-item online">在线: ${summary.online}</div>
            <div class="summary-item offline">离线: ${summary.offline}</div>
            <div class="summary-item timeout">超时: ${summary.timeout}</div>
        `;
    }
    
    // 显示详细状态
    statusList.innerHTML = '';
    statuses.forEach(status => {
        const statusItem = document.createElement('div');
        statusItem.className = `status-item ${status.status}`;
        
        const latencyText = status.latency > 0 ? `${status.latency}ms` : '-';
        const errorText = status.error ? `错误: ${status.error}` : '';
        
        statusItem.innerHTML = `
            <div class="node-name">${status.node.name}</div>
            <div class="node-server">${status.node.server}:${status.node.port}</div>
            <div class="node-status">${getStatusText(status.status)}</div>
            <div class="node-latency">${latencyText}</div>
            ${errorText ? `<div class="node-error">${errorText}</div>` : ''}
        `;
        
        statusList.appendChild(statusItem);
    });
}

// 获取状态文本
function getStatusText(status) {
    const statusMap = {
        'online': '在线',
        'offline': '离线',
        'timeout': '超时'
    };
    return statusMap[status] || status;
}

// 复制订阅链接
async function copySubscriptionUrl() {
    const urlInput = document.getElementById('subscriptionUrl');
    if (!urlInput.value) {
        showMessage('没有可复制的链接', 'error');
        return;
    }
    
    try {
        await navigator.clipboard.writeText(urlInput.value);
        showMessage('链接已复制到剪贴板', 'success');
        
        // 视觉反馈
        const copyBtn = document.getElementById('copyUrlBtn');
        const originalText = copyBtn.textContent;
        copyBtn.textContent = '已复制';
        copyBtn.classList.add('copied');
        
        setTimeout(() => {
            copyBtn.textContent = originalText;
            copyBtn.classList.remove('copied');
        }, 2000);
    } catch (error) {
        // 降级到选择文本
        urlInput.select();
        urlInput.setSelectionRange(0, 99999);
        showMessage('请手动复制链接', 'info');
    }
}

// 切换配置预览
function toggleConfigPreview() {
    const configContent = document.getElementById('configContent');
    const toggleBtn = document.getElementById('toggleConfig');
    
    if (configContent.style.display === 'none') {
        configContent.style.display = 'block';
        toggleBtn.textContent = '隐藏配置';
    } else {
        configContent.style.display = 'none';
        toggleBtn.textContent = '显示配置';
    }
}

// 下载配置文件
function downloadConfig() {
    if (!window.currentConfig || !window.currentConfig.content) {
        showMessage('没有可下载的配置', 'error');
        return;
    }
    
    const blob = new Blob([window.currentConfig.content], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = window.currentConfig.filename || 'clash_config.yaml';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    showMessage('配置文件下载开始', 'success');
}

// 从URL提取文件名
function extractFilenameFromUrl(url) {
    if (!url) return 'clash_config.yaml';
    const parts = url.split('/');
    return parts[parts.length - 1] || 'clash_config.yaml';
}

// 显示消息
function showMessage(text, type = 'info') {
    const messageEl = document.getElementById('message');
    messageEl.textContent = text;
    messageEl.className = `message-toast ${type} show`;
    
    // 自动隐藏
    setTimeout(() => {
        messageEl.classList.remove('show');
    }, 3000);
}

// 重置订阅
async function resetSubscription() {
    // 确认对话框
    if (!confirm('⚠️ 确定要重置所有订阅链接吗？\n\n这将删除您的所有历史订阅文件，此操作不可恢复！')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('jwt_token');
        showMessage('正在重置订阅...', 'info');
        
        const response = await fetch('/api/reset-subscription', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showMessage(`✅ ${data.message}`, 'success');
            
            // 清空当前显示的订阅链接
            const subscriptionUrl = document.getElementById('subscriptionUrl');
            if (subscriptionUrl) {
                subscriptionUrl.value = '';
            }
            
            // 隐藏结果区域
            const resultSection = document.getElementById('resultSection');
            if (resultSection) {
                resultSection.style.display = 'none';
            }
            
            // 清空配置内容
            const configContent = document.getElementById('configContent');
            if (configContent) {
                configContent.textContent = '';
            }
            
            // 清空全局配置缓存
            window.currentConfig = null;
            
        } else {
            showMessage(data.message || '重置失败', 'error');
        }
    } catch (error) {
        console.error('重置订阅错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    }
}

// 处理网络错误
window.addEventListener('unhandledrejection', function(event) {
    if (event.reason && event.reason.name === 'TypeError' && event.reason.message.includes('fetch')) {
        showMessage('网络连接错误，请检查网络状态', 'error');
    }
});

