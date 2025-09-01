// frontend/init.js
document.addEventListener('DOMContentLoaded', function() {
    // 检查是否已经初始化
    checkInitStatus();
    
    // 绑定事件
    setupEventListeners();
});

// 检查初始化状态
async function checkInitStatus() {
    try {
        const response = await fetch('/api/init', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                adminUsername: 'check',
                adminPassword: 'check'
            })
        });

        if (response.status === 403) {
            // 系统未初始化，这是正常的
            return;
        }

        const data = await response.json();
        if (!data.success && data.message === '系统已经初始化') {
            // 系统已经初始化，重定向到登录页
            showMessage('系统已经初始化，正在跳转到登录页...', 'info');
            setTimeout(() => {
                window.location.href = '/login';
            }, 2000);
        }
    } catch (error) {
        console.error('检查初始化状态错误:', error);
    }
}

// 设置事件监听器
function setupEventListeners() {
    // 初始化表单提交
    document.getElementById('initForm').addEventListener('submit', handleInit);
    
    // 密码确认验证
    const confirmPassword = document.getElementById('confirmPassword');
    const adminPassword = document.getElementById('adminPassword');
    
    confirmPassword.addEventListener('input', function() {
        if (this.value !== adminPassword.value) {
            this.setCustomValidity('密码不匹配');
        } else {
            this.setCustomValidity('');
        }
    });
    
    adminPassword.addEventListener('input', function() {
        if (confirmPassword.value && confirmPassword.value !== this.value) {
            confirmPassword.setCustomValidity('密码不匹配');
        } else {
            confirmPassword.setCustomValidity('');
        }
    });
}

// 处理系统初始化
async function handleInit(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const adminUsername = formData.get('adminUsername').trim();
    const adminPassword = formData.get('adminPassword');
    const confirmPassword = formData.get('confirmPassword');
    
    if (!adminUsername || !adminPassword || !confirmPassword) {
        showMessage('请填写完整信息', 'error');
        return;
    }
    
    if (adminUsername.length < 3) {
        showMessage('管理员用户名至少需要3个字符', 'error');
        return;
    }
    
    if (adminPassword.length < 6) {
        showMessage('密码至少需要6个字符', 'error');
        return;
    }
    
    if (adminPassword !== confirmPassword) {
        showMessage('两次输入的密码不一致', 'error');
        return;
    }
    
    const initBtn = document.querySelector('.init-btn');
    
    try {
        // 显示加载状态
        initBtn.classList.add('loading');
        initBtn.disabled = true;
        showMessage('正在初始化系统...', 'info');
        
        const response = await fetch('/api/init', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                adminUsername: adminUsername,
                adminPassword: adminPassword
            })
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showMessage('🎉 系统初始化成功！正在跳转到登录页...', 'success');
            // 清空表单
            document.getElementById('initForm').reset();
            // 延迟跳转
            setTimeout(() => {
                window.location.href = '/login';
            }, 3000);
        } else {
            showMessage(data.message || '初始化失败', 'error');
        }
    } catch (error) {
        console.error('初始化错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        // 恢复按钮状态
        initBtn.classList.remove('loading');
        initBtn.disabled = false;
    }
}

// 显示消息
function showMessage(text, type = 'info') {
    const messageEl = document.getElementById('message');
    messageEl.textContent = text;
    messageEl.className = `message ${type}`;
    messageEl.style.display = 'block';
}

// 清除消息
function clearMessage() {
    const messageEl = document.getElementById('message');
    messageEl.style.display = 'none';
    messageEl.textContent = '';
    messageEl.className = 'message';
}
