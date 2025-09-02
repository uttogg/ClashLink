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
    
    // YAML 编辑器功能
    document.getElementById('editConfig').addEventListener('click', startEditConfig);
    document.getElementById('saveConfig').addEventListener('click', saveEditedConfig);
    document.getElementById('cancelEdit').addEventListener('click', cancelEditConfig);
    
    // YAML 编辑器实时更新
    document.getElementById('yamlEditor').addEventListener('input', updateEditorStatus);
    
    // 默认配置管理
    document.getElementById('saveDefaultConfig').addEventListener('click', saveDefaultConfig);
    document.getElementById('loadDefaultConfig').addEventListener('click', loadDefaultConfig);
    document.getElementById('resetDefaultConfig').addEventListener('click', resetDefaultConfig);
    
    // 页面加载时自动加载默认配置
    loadDefaultConfig();
    
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
    const configName = document.getElementById('configName').value.trim() || 
                      document.getElementById('defaultConfigName').value.trim() || 'ClashLink配置';
    
    // 获取自定义配置参数
    const mixedPort = parseInt(document.getElementById('mixedPort').value) || 7890;
    const controllerPort = parseInt(document.getElementById('controllerPort').value) || 9090;
    const allowLan = document.getElementById('allowLan').checked;
    const logLevel = document.getElementById('logLevel').value;
    const dnsMode = document.getElementById('dnsMode').value;
    const enableIPv6 = document.getElementById('enableIPv6').checked;
    const customRules = document.getElementById('customRules').value.trim();
    
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
                configName: configName,
                mixedPort: mixedPort,
                controllerPort: controllerPort,
                allowLan: allowLan,
                logLevel: logLevel,
                dnsMode: dnsMode,
                enableIPv6: enableIPv6,
                customRules: customRules
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
            
            // 清空编辑器
            const yamlEditor = document.getElementById('yamlEditor');
            if (yamlEditor) {
                yamlEditor.value = '';
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

// 开始编辑配置
function startEditConfig() {
    const configContent = document.getElementById('configContent');
    const configEditor = document.getElementById('configEditor');
    const yamlEditor = document.getElementById('yamlEditor');
    
    // 检查是否有配置内容
    if (!window.currentConfig || !window.currentConfig.content) {
        showMessage('请先生成订阅配置', 'error');
        return;
    }
    
    // 切换到编辑模式
    configContent.style.display = 'none';
    configEditor.style.display = 'block';
    
    // 设置编辑器内容
    yamlEditor.value = window.currentConfig.content;
    
    // 更新按钮显示
    document.getElementById('toggleConfig').style.display = 'none';
    document.getElementById('editConfig').style.display = 'none';
    document.getElementById('saveConfig').style.display = 'inline-block';
    document.getElementById('cancelEdit').style.display = 'inline-block';
    
    // 更新状态
    updateEditorStatus();
    
    // 聚焦编辑器
    yamlEditor.focus();
    
    showMessage('进入编辑模式，您可以直接修改 YAML 配置', 'info');
}

// 取消编辑
function cancelEditConfig() {
    const configContent = document.getElementById('configContent');
    const configEditor = document.getElementById('configEditor');
    
    // 切换回预览模式
    configEditor.style.display = 'none';
    if (configContent.textContent) {
        configContent.style.display = 'block';
    }
    
    // 恢复按钮显示
    document.getElementById('toggleConfig').style.display = 'inline-block';
    document.getElementById('editConfig').style.display = 'inline-block';
    document.getElementById('saveConfig').style.display = 'none';
    document.getElementById('cancelEdit').style.display = 'none';
    
    showMessage('已取消编辑', 'info');
}

// 保存编辑的配置
async function saveEditedConfig() {
    const yamlEditor = document.getElementById('yamlEditor');
    const editedContent = yamlEditor.value.trim();
    
    if (!editedContent) {
        showMessage('配置内容不能为空', 'error');
        return;
    }
    
    try {
        const token = localStorage.getItem('jwt_token');
        showMessage('正在保存配置...', 'info');
        
        const response = await fetch('/api/save-config', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                configContent: editedContent,
                filename: window.currentConfig ? window.currentConfig.filename : null
            })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showMessage('✅ 配置保存成功', 'success');
            
            // 更新全局配置缓存
            if (window.currentConfig) {
                window.currentConfig.content = editedContent;
            }
            
            // 更新预览内容
            const configContent = document.getElementById('configContent');
            configContent.textContent = editedContent;
            
            // 更新订阅链接
            if (data.subscriptionUrl) {
                const subscriptionUrl = document.getElementById('subscriptionUrl');
                if (subscriptionUrl) {
                    subscriptionUrl.value = data.subscriptionUrl;
                }
            }
            
            // 退出编辑模式
            cancelEditConfig();
            
        } else {
            showMessage(data.message || '保存失败', 'error');
        }
    } catch (error) {
        console.error('保存配置错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    }
}

// 更新编辑器状态
function updateEditorStatus() {
    const yamlEditor = document.getElementById('yamlEditor');
    const editorStatus = document.getElementById('editorStatus');
    const lineCount = document.getElementById('lineCount');
    
    if (!yamlEditor || !editorStatus || !lineCount) return;
    
    const content = yamlEditor.value;
    const lines = content.split('\n').length;
    
    lineCount.textContent = `${lines} 行`;
    
    // 简单的YAML语法检查
    try {
        // 检查基本的YAML语法
        if (content.trim() === '') {
            editorStatus.textContent = '配置为空';
            editorStatus.className = 'status-text error';
        } else if (content.includes('proxies:')) {
            editorStatus.textContent = '配置格式正常';
            editorStatus.className = 'status-text success';
        } else {
            editorStatus.textContent = '请确保包含 proxies 部分';
            editorStatus.className = 'status-text error';
        }
    } catch (e) {
        editorStatus.textContent = '配置格式错误';
        editorStatus.className = 'status-text error';
    }
}

// 保存默认配置
function saveDefaultConfig() {
    const configName = document.getElementById('defaultConfigName').value.trim();
    const customRules = document.getElementById('customRules').value.trim();
    
    if (!configName) {
        showMessage('请输入配置名称', 'error');
        return;
    }
    
    // 收集当前的高级配置
    const advancedConfig = {
        mixedPort: parseInt(document.getElementById('mixedPort').value) || 7890,
        controllerPort: parseInt(document.getElementById('controllerPort').value) || 9090,
        allowLan: document.getElementById('allowLan').checked,
        logLevel: document.getElementById('logLevel').value,
        dnsMode: document.getElementById('dnsMode').value,
        enableIPv6: document.getElementById('enableIPv6').checked,
        configName: configName,
        customRules: customRules
    };
    
    // 保存到localStorage
    localStorage.setItem('clashlink_default_config', JSON.stringify(advancedConfig));
    
    showMessage('✅ 默认配置已保存', 'success');
}

// 加载默认配置
function loadDefaultConfig() {
    const savedConfig = localStorage.getItem('clashlink_default_config');
    
    if (savedConfig) {
        try {
            const config = JSON.parse(savedConfig);
            
            // 应用保存的配置
            if (config.mixedPort) document.getElementById('mixedPort').value = config.mixedPort;
            if (config.controllerPort) document.getElementById('controllerPort').value = config.controllerPort;
            if (config.allowLan !== undefined) document.getElementById('allowLan').checked = config.allowLan;
            if (config.logLevel) document.getElementById('logLevel').value = config.logLevel;
            if (config.dnsMode) document.getElementById('dnsMode').value = config.dnsMode;
            if (config.enableIPv6 !== undefined) document.getElementById('enableIPv6').checked = config.enableIPv6;
            if (config.configName) document.getElementById('defaultConfigName').value = config.configName;
            if (config.customRules) document.getElementById('customRules').value = config.customRules;
            
            showMessage('📥 默认配置已加载', 'info');
        } catch (e) {
            console.error('加载默认配置失败:', e);
            showMessage('加载默认配置失败', 'error');
        }
    } else {
        // 设置默认值
        document.getElementById('defaultConfigName').value = 'ClashLink配置';
        showMessage('📝 使用默认配置', 'info');
    }
}

// 重置默认配置
function resetDefaultConfig() {
    if (confirm('确定要重置所有默认配置吗？这将恢复出厂设置。')) {
        localStorage.removeItem('clashlink_default_config');
        
        // 重置所有字段到默认值
        document.getElementById('mixedPort').value = '7890';
        document.getElementById('controllerPort').value = '9090';
        document.getElementById('allowLan').checked = true;
        document.getElementById('logLevel').value = 'info';
        document.getElementById('dnsMode').value = 'fake-ip';
        document.getElementById('enableIPv6').checked = false;
        document.getElementById('defaultConfigName').value = 'ClashLink配置';
        document.getElementById('customRules').value = '';
        
        showMessage('🔄 默认配置已重置', 'info');
    }
}

// 处理网络错误
window.addEventListener('unhandledrejection', function(event) {
    if (event.reason && event.reason.name === 'TypeError' && event.reason.message.includes('fetch')) {
        showMessage('网络连接错误，请检查网络状态', 'error');
    }
});

