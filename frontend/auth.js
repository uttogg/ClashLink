// frontend/auth.js
document.addEventListener('DOMContentLoaded', function() {
    // 检查是否已经登录
    checkExistingLogin();
    
    // 绑定事件
    setupEventListeners();
});

// 检查现有登录状态
function checkExistingLogin() {
    const token = localStorage.getItem('jwt_token');
    if (token) {
        // 验证token是否有效
        if (isTokenValid(token)) {
            window.location.href = '/app';
            return;
        } else {
            // token无效，清除
            localStorage.removeItem('jwt_token');
        }
    }
}

// 简单的token有效性检查（检查是否过期）
function isTokenValid(token) {
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        const now = Math.floor(Date.now() / 1000);
        return payload.exp > now;
    } catch (e) {
        return false;
    }
}

// 设置事件监听器
function setupEventListeners() {
    // 标签切换
    const tabBtns = document.querySelectorAll('.tab-btn');
    const forms = document.querySelectorAll('.auth-form');
    
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tab = btn.dataset.tab;
            
            // 更新标签状态
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            // 更新表单显示
            forms.forEach(form => {
                form.classList.remove('active');
                if (form.id === tab + 'Form') {
                    form.classList.add('active');
                }
            });
            
            // 清除消息
            clearMessage();
        });
    });
    
    // 登录表单提交
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    
    // 注册表单提交
    document.getElementById('registerForm').addEventListener('submit', handleRegister);
    
    // 密码确认验证
    const confirmPassword = document.getElementById('confirmPassword');
    const registerPassword = document.getElementById('registerPassword');
    
    confirmPassword.addEventListener('input', function() {
        if (this.value !== registerPassword.value) {
            this.setCustomValidity('密码不匹配');
        } else {
            this.setCustomValidity('');
        }
    });
    
    registerPassword.addEventListener('input', function() {
        if (confirmPassword.value && confirmPassword.value !== this.value) {
            confirmPassword.setCustomValidity('密码不匹配');
        } else {
            confirmPassword.setCustomValidity('');
        }
    });
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const username = formData.get('username').trim();
    const password = formData.get('password');
    
    if (!username || !password) {
        showMessage('请填写完整信息', 'error');
        return;
    }
    
    try {
        // 添加加载动画
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.classList.add('loading');
        submitBtn.disabled = true;
        showMessage('正在登录...', 'info');
        
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok && data.token) {
            localStorage.setItem('jwt_token', data.token);
            showMessage('登录成功，正在跳转...', 'success');
            setTimeout(() => {
                window.location.href = '/app';
            }, 1000);
        } else {
            showMessage(data.message || '登录失败', 'error');
        }
    } catch (error) {
        console.error('登录错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        // 恢复按钮状态
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.classList.remove('loading');
        submitBtn.disabled = false;
    }
}

// 处理注册
async function handleRegister(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const username = formData.get('username').trim();
    const password = formData.get('password');
    const confirmPassword = formData.get('confirmPassword');
    
    if (!username || !password || !confirmPassword) {
        showMessage('请填写完整信息', 'error');
        return;
    }
    
    if (username.length < 3) {
        showMessage('用户名至少需要3个字符', 'error');
        return;
    }
    
    if (password.length < 6) {
        showMessage('密码至少需要6个字符', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showMessage('两次输入的密码不一致', 'error');
        return;
    }
    
    try {
        // 添加加载动画
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.classList.add('loading');
        submitBtn.disabled = true;
        showMessage('正在注册...', 'info');
        
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage('注册成功！请登录', 'success');
            // 切换到登录表单
            document.querySelector('.tab-btn[data-tab="login"]').click();
            // 清空注册表单
            document.getElementById('registerForm').reset();
            // 填入用户名到登录表单
            document.getElementById('loginUsername').value = username;
        } else {
            showMessage(data.message || response.statusText, 'error');
        }
    } catch (error) {
        console.error('注册错误:', error);
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        // 恢复按钮状态
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.classList.remove('loading');
        submitBtn.disabled = false;
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

// 导出函数供其他脚本使用
window.AuthUtils = {
    getToken: () => localStorage.getItem('jwt_token'),
    setToken: (token) => localStorage.setItem('jwt_token', token),
    removeToken: () => localStorage.removeItem('jwt_token'),
    isTokenValid: isTokenValid,
    logout: () => {
        localStorage.removeItem('jwt_token');
        window.location.href = '/login';
    }
};

