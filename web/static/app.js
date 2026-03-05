let login_token;

function logout() {
    const logoutBtn = document.querySelector('.Btn');
    if (logoutBtn) logoutBtn.classList.add('loading');
    if (confirm('确定要退出登录吗？')) {
        // 标记登出并清理本地存储/会话/cookie
        localStorage.setItem('justLoggedOut', 'true');
        const cleanup = () => {
            clearAuthState();
            sessionStorage.clear();
            window.location.href = '/';
        };

        const token = getAuthToken();
        const logoutRequest = token
            ? fetch(API_BASE + "/auth/logout", {
                  method: "POST",
                  headers: getAuthHeaders({ "Content-Type": "application/json" }),
              }).catch(() => {})
            : Promise.resolve();

        logoutRequest.finally(cleanup);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const sideMenu = document.querySelector("aside");
    const themeToggler = document.querySelector(".theme-toggler");
    const nextDay = document.getElementById('nextDay');
    const prevDay = document.getElementById('prevDay');
    const timetable = document.querySelector('.timetable');
    const timetableTitle = document.querySelector('.timetable div h2');
    const tableBody = document.querySelector('table tbody');
    const header = document.querySelector('header');

    // Profile button toggle for side menu
    // profileBtn.onclick = function () {
    //     sideMenu.classList.toggle('active');
    // }

    // Scroll event to remove side menu and add/remove header active class
    window.onscroll = () => {
        if (sideMenu) {
            sideMenu.classList.remove('active');
        }
        if (!header) return;
        if (window.scrollY > 0) {
            header.classList.add('active');
        } else {
            header.classList.remove('active');
        }
    }

    // Theme toggle function
    const applySavedTheme = () => {
        if (!themeToggler) return;
        const isDarkMode = localStorage.getItem('dark-theme') === 'true';
        if (isDarkMode) {
            document.body.classList.add('dark-theme');
            themeToggler.querySelector('span:nth-child(2)').classList.add('active');
            themeToggler.querySelector('span:nth-child(1)').classList.remove('active');
        } else {
            document.body.classList.remove('dark-theme');
            themeToggler.querySelector('span:nth-child(2)').classList.remove('active');
            themeToggler.querySelector('span:nth-child(1)').classList.add('active');
        }
    }

    // Set the initial theme based on localStorage
    applySavedTheme();

    // Toggle theme function
    if (themeToggler) themeToggler.onclick = function () {
        // Toggle dark theme class on body
        document.body.classList.toggle('dark-theme');

        // Toggle active class on the theme toggler spans
        themeToggler.querySelector('span:nth-child(1)').classList.toggle('active');
        themeToggler.querySelector('span:nth-child(2)').classList.toggle('active');

        // Save the theme preference in localStorage
        localStorage.setItem('dark-theme', document.body.classList.contains('dark-theme'));
    }

    // Function to set timetable data
    let setData = (day) => {
        if (!tableBody || !timetableTitle) return;
        tableBody.innerHTML = '';
        let daylist = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
        timetableTitle.innerHTML = daylist[day];

        // Define subjects for each day (you might need to update this with real data)
        let daySchedule = [];
        switch (day) {
            case 0: daySchedule = Sunday; break;
            case 1: daySchedule = Monday; break;
            case 2: daySchedule = Tuesday; break;
            case 3: daySchedule = Wednesday; break;
            case 4: daySchedule = Thursday; break;
            case 5: daySchedule = Friday; break;
            case 6: daySchedule = Saturday; break;
        }

        // Append timetable data to table
        daySchedule.forEach(sub => {
            const tr = document.createElement('tr');
            const trContent = `
                <td>${sub.time}</td>
                <td>${sub.roomNumber}</td>
                <td>${sub.subject}</td>
                <td>${sub.type}</td>
            `;
            tr.innerHTML = trContent;
            tableBody.appendChild(tr);
        });
    }

    // Get current day and set timetable on page load
    let now = new Date();
    let today = now.getDay();  // Get current day (0 - 6)
    let day = today;  // To prevent today value from changing

    // Function to toggle timetable visibility
    function timeTableAll() {
        const timetableById = document.getElementById('timetable');
        if (!timetableById || !timetableTitle) return;
        timetableById.classList.toggle('active');
        setData(today);
        timetableTitle.innerHTML = "Today's Timetable";
    }

    // Event listeners for next and previous day buttons
    if (nextDay) {
        nextDay.onclick = function () {
            day <= 5 ? day++ : day = 0;
            setData(day);
        }
    }

    if (prevDay) {
        prevDay.onclick = function () {
            day >= 1 ? day-- : day = 6;
            setData(day);
        }
    }

    if (timetable && tableBody && timetableTitle) {
        setData(day);
        timetableTitle.innerHTML = "Today's Timetable";
    }
});


const PUBLIC_PATHS = new Set(["/", "/login", "/register", "/register_result"]);

function isPublicPath(pathname) {
    return PUBLIC_PATHS.has(pathname);
}

function clearAuthState() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('userEmail');
    localStorage.removeItem('userId');
    localStorage.removeItem('userName');
    document.cookie = "token=; path=/; max-age=0";
}

function redirectToLogin() {
    window.location.href = "/";
    window.location.href = "/login";
}

function getCookieValue(name) {
    const matches = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()[\]\\/+^])/g, '\\$1') + '=([^;]*)'));
    return matches ? decodeURIComponent(matches[1]) : '';
}

function getAuthToken() {
    return localStorage.getItem('authToken') || getCookieValue('token');
}

function getUserProfile(options = {}) {
    const token = getAuthToken();
    const requireAuth = options.requireAuth === true;

    if (!token) {
        if (requireAuth && !isPublicPath(window.location.pathname)) {
            redirectToLogin();
        }
        return;
    }

    return fetch("/api/v1/user/profile", {
        method: "GET",
        headers: {
            "Authorization": "Bearer " + token,
            "Accept": "application/json"
        }
    })
    .then(res => {
        if (!res.ok) throw new Error("HTTP error " + res.status);
        return res.json();
    })
    .then(data => {
        console.log("用户资料：", data);

        const emailElement = document.getElementById("displayed_email");
        if (emailElement && data && data.data && data.data.email) {
            const email = data.data.email;
            const formattedEmail = email.replace('@', '<wbr>@');
            emailElement.innerHTML = formattedEmail;
            console.log("邮箱已设置：" + data.data.email);
        } else if (emailElement) {
            console.error("无法获取邮箱，响应数据：", data);
            emailElement.innerHTML = "邮箱获取失败";
        }
        if (data && data.data && data.data.avatar_url) {
            loadUserAvatar(data.data.avatar_url);
        }
        return data;
    })
    .catch(err => {
        console.error("获取用户资料失败:", err);
        if (requireAuth && !isPublicPath(window.location.pathname)) {
            clearAuthState();
            redirectToLogin();
        }
    });
}

function loadUserAvatar(avatarUrl) {
    const token = getAuthToken();
    if (!token) return;
    const avatarImages = document.querySelectorAll('img.profile-avatar');
    if (!avatarImages || avatarImages.length === 0) return;

    const url = avatarUrl || "/api/v1/user/avatar";
    fetch(url, {
        method: "GET",
        headers: {
            "Authorization": "Bearer " + token
        }
    })
    .then(res => {
        if (!res.ok) {
            throw new Error("avatar missing");
        }
        return res.blob();
    })
    .then(blob => {
        if (!blob || blob.size === 0) return;
        const objectUrl = URL.createObjectURL(blob);
        avatarImages.forEach(img => {
            if (img.dataset.objectUrl) {
                URL.revokeObjectURL(img.dataset.objectUrl);
            }
            img.dataset.objectUrl = objectUrl;
            img.src = objectUrl;
        });
    })
    .catch(() => {});
}

function initAuthGate() {
    if (isPublicPath(window.location.pathname)) return;
    getUserProfile({ requireAuth: true });
}

document.addEventListener("DOMContentLoaded", initAuthGate);

const API_BASE = '/api/v1'

function isSuccess(resp) {
    return resp && (resp.code === 0 || resp.code === 200)
}

function parseJsonSafe(res) {
    return res.text().then(text => {
        if (!text) return { _empty: true };
        try {
            return JSON.parse(text);
        } catch (err) {
            return { _rawText: text };
        }
    });
}

function getErrorMessage(data, fallback) {
    if (data && typeof data === 'object') {
        if (data.message) return data.message;
        if (data.msg) return data.msg;
        if (data.error) return data.error;
        if (data._rawText) return data._rawText.trim() || fallback;
    }
    return fallback;
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
}

function escapeHTML(value) {
    if (value === null || value === undefined) return '';
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function getAuthHeaders(extra = {}) {
    const token = getAuthToken();
    const headers = { ...extra };
    if (token) headers.Authorization = 'Bearer ' + token;
    return headers;
}

function changePassword() {
    const result = document.getElementById('passwordResult');
    if (result) result.innerText = '';

    if (!getAuthToken()) {
        if (result) result.innerText = 'Please sign in again.';
        return;
    }

    const oldPass = document.getElementById('currentpass') && document.getElementById('currentpass').value;
    const newPass = document.getElementById('newpass') && document.getElementById('newpass').value;
    const confirmPass = document.getElementById('confirmpass') && document.getElementById('confirmpass').value;

    if (!oldPass || !newPass) {
        if (result) result.innerText = 'Please fill in all password fields.';
        return;
    }

    if (newPass.length < 6) {
        if (result) result.innerText = 'New password must be at least 6 characters.';
        return;
    }

    if (newPass !== confirmPass) {
        if (result) result.innerText = 'Passwords do not match.';
        return;
    }

    fetch(API_BASE + '/user/password', {
        method: 'PUT',
        headers: getAuthHeaders({ 'Content-Type': 'application/json', 'Accept': 'application/json' }),
        body: JSON.stringify({ old_password: oldPass, new_password: newPass })
    })
        .then(res => parseJsonSafe(res).then(data => ({ res, data })))
        .then(({ res, data }) => {
            if (!res.ok) {
                throw new Error(getErrorMessage(data, 'Update failed'));
            }
            if (result) result.innerText = 'Password updated.';
            const oldInput = document.getElementById('currentpass');
            const newInput = document.getElementById('newpass');
            const confirmInput = document.getElementById('confirmpass');
            if (oldInput) oldInput.value = '';
            if (newInput) newInput.value = '';
            if (confirmInput) confirmInput.value = '';
        })
        .catch(err => {
            console.error(err);
            if (result) result.innerText = err.message || 'Update failed';
        });
}

function renderFiles(items) {
    const filesBody = document.getElementById('filesBody');
    if (!filesBody) return;

    if (!items || items.length === 0) {
        filesBody.innerHTML = '<tr><td colspan="5" class="text-muted">No files in vault.</td></tr>';
        return;
    }

    filesBody.innerHTML = items.map(file => {
        const created = file.created_at ? new Date(file.created_at).toLocaleString() : '-';
        return `
        <tr>
            <td>${escapeHTML(file.filename || '-')}</td>
            <td>${formatSize(Number(file.size || 0))}</td>
            <td>${escapeHTML(created)}</td>
            <td>${escapeHTML(file.description || '-')}</td>
            <td>
                <button class="file-action" data-action="download" data-id="${file.id}" data-name="${escapeHTML(file.filename || 'file')}">Download</button>
                <button class="file-action update" data-action="update" data-id="${file.id}">Update</button>
                <button class="file-action delete" data-action="delete" data-id="${file.id}">Delete</button>
            </td>
        </tr>`;
    }).join('');
}

function loadSecureFiles() {
    const filesBody = document.getElementById('filesBody');
    if (!filesBody) return Promise.resolve();

    const token = getAuthToken();
    const refreshBtn = document.getElementById('refreshFilesBtn');
    if (!token) {
        filesBody.innerHTML = '<tr><td colspan="5" class="text-muted">Sign in to view protected files.</td></tr>';
        if (refreshBtn) {
            refreshBtn.disabled = true;
            refreshBtn.title = 'Sign in to refresh protected files.';
        }
        return Promise.resolve();
    }
    if (refreshBtn) {
        refreshBtn.disabled = false;
        refreshBtn.title = '';
    }

    filesBody.innerHTML = '<tr><td colspan="5" class="text-muted">Loading files...</td></tr>';
    return fetch(API_BASE + '/files?page=1&size=50', {
        method: 'GET',
        headers: getAuthHeaders({ Accept: 'application/json' })
    })
        .then(res => {
            return parseJsonSafe(res).then(data => {
                if (!res.ok) {
                    if (res.status === 401) {
                        clearAuthState();
                        redirectToLogin();
                        throw new Error('Session expired. Please log in again.');
                    }
                    throw new Error(getErrorMessage(data, 'HTTP error ' + res.status));
                }
                return data;
            });
        })
        .then(data => {
            if (!isSuccess(data) || !data.data) {
                throw new Error(data.message || 'Failed to load files');
            }
            renderFiles(data.data.items || []);
        })
        .catch(err => {
            filesBody.innerHTML = `<tr><td colspan="5" class="text-muted">Failed to load: ${escapeHTML(err.message || err)}</td></tr>`;
        });
}

function downloadSecureFile(fileId, filename) {
    return fetch(API_BASE + '/files/download/' + fileId, {
        method: 'GET',
        headers: getAuthHeaders()
    })
        .then(res => {
            if (!res.ok) throw new Error('Download failed');
            return res.blob();
        })
        .then(blob => {
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename || 'download';
            document.body.appendChild(a);
            a.click();
            a.remove();
            URL.revokeObjectURL(url);
        });
}

document.addEventListener('DOMContentLoaded', () => {
    const uploadBtn = document.getElementById('uploadBtn');
    const uploadPublicBtn = document.getElementById('uploadPublicBtn');
    const fileInput = document.getElementById('fileInput');
    const fileDesc = document.getElementById('fileDesc');
    const uploadResult = document.getElementById('uploadResult');
    const refreshFilesBtn = document.getElementById('refreshFilesBtn');
    const filesBody = document.getElementById('filesBody');

    if (!uploadBtn || !fileInput || !uploadResult || !filesBody) return;

    loadSecureFiles();

    const runUpload = (endpoint, requireAuth, buttonEl) => {
        uploadResult.innerText = '';
        if (!fileInput.files || fileInput.files.length === 0) {
            uploadResult.innerText = 'Please choose a file first.';
            return;
        }
        if (requireAuth) {
            const token = getAuthToken();
            if (!token) {
                uploadResult.innerText = 'Session expired. Please log in again.';
                redirectToLogin();
                return;
            }
        }

        const fd = new FormData();
        fd.append('file', fileInput.files[0]);
        fd.append('description', fileDesc ? fileDesc.value.trim() : '');

        if (buttonEl) {
            buttonEl.disabled = true;
            buttonEl.innerText = 'Uploading...';
        }

        fetch(API_BASE + endpoint, {
            method: 'POST',
            headers: requireAuth ? getAuthHeaders() : {},
            body: fd
        })
            .then(res => {
                return parseJsonSafe(res).then(data => {
                    if (!res.ok) {
                        if (res.status === 401) {
                            clearAuthState();
                            redirectToLogin();
                            throw new Error('Session expired. Please log in again.');
                        }
                        throw new Error(getErrorMessage(data, 'Upload failed'));
                    }
                    return data;
                });
            })
            .then(data => {
                if (!isSuccess(data)) throw new Error(data.message || 'Upload failed');
                uploadResult.innerText = 'Upload completed.';
                fileInput.value = '';
                if (fileDesc) fileDesc.value = '';
                if (requireAuth) {
                    return loadSecureFiles();
                }
            })
            .catch(err => {
                uploadResult.innerText = 'Upload failed: ' + (err.message || err);
            })
            .finally(() => {
                if (buttonEl) {
                    buttonEl.disabled = false;
                    buttonEl.innerText = buttonEl.dataset.label || 'Upload';
                }
            });
    };

    const runUpdate = (fileId, buttonEl) => {
        uploadResult.innerText = '';
        const hasFile = fileInput.files && fileInput.files.length > 0;
        const descValue = fileDesc ? fileDesc.value : '';
        if (!hasFile && descValue === '') {
            uploadResult.innerText = 'Choose a file or enter a description to update.';
            return;
        }

        const fd = new FormData();
        if (hasFile) {
            fd.append('file', fileInput.files[0]);
        }
        if (fileDesc) {
            fd.append('description', descValue);
        }

        if (buttonEl) {
            buttonEl.disabled = true;
            buttonEl.innerText = 'Updating...';
        }

        fetch(API_BASE + '/files/' + fileId, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: fd
        })
            .then(res => {
                return parseJsonSafe(res).then(data => {
                    if (!res.ok) {
                        if (res.status === 401) {
                            clearAuthState();
                            redirectToLogin();
                            throw new Error('Session expired. Please log in again.');
                        }
                        throw new Error(getErrorMessage(data, 'Update failed'));
                    }
                    return data;
                });
            })
            .then(data => {
                if (!isSuccess(data)) throw new Error(data.message || 'Update failed');
                uploadResult.innerText = 'Update completed.';
                fileInput.value = '';
                if (fileDesc) fileDesc.value = '';
                return loadSecureFiles();
            })
            .catch(err => {
                uploadResult.innerText = 'Update failed: ' + (err.message || err);
            })
            .finally(() => {
                if (buttonEl) {
                    buttonEl.disabled = false;
                    buttonEl.innerText = buttonEl.dataset.label || 'Update';
                }
            });
    };

    uploadBtn.dataset.label = 'Upload Securely';
    uploadBtn.addEventListener('click', () => {
        runUpload('/files/upload', true, uploadBtn);
    });

    if (uploadPublicBtn) {
        uploadPublicBtn.dataset.label = 'Upload Publicly';
        uploadPublicBtn.addEventListener('click', () => {
            runUpload('/files/public/upload', false, uploadPublicBtn);
        });
    }

    if (refreshFilesBtn) {
        refreshFilesBtn.addEventListener('click', () => {
            loadSecureFiles();
        });
    }

    filesBody.addEventListener('click', (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;
        const action = target.dataset.action;
        const fileId = target.dataset.id;
        if (!action || !fileId) return;

        if (action === 'download') {
            const filename = target.dataset.name || 'download';
            downloadSecureFile(fileId, filename)
                .catch(err => {
                    uploadResult.innerText = 'Download failed: ' + (err.message || err);
                });
            return;
        }

        if (action === 'update') {
            if (!confirm('Update this file in vault?')) return;
            const buttonEl = target;
            if (!buttonEl.dataset.label) {
                buttonEl.dataset.label = buttonEl.innerText || 'Update';
            }
            runUpdate(fileId, buttonEl);
            return;
        }

        if (action === 'delete') {
            if (!confirm('Delete this file from vault?')) return;
            fetch(API_BASE + '/files/' + fileId, {
                method: 'DELETE',
                headers: getAuthHeaders()
            })
                .then(res => {
                    if (res.status === 204) return;
                    throw new Error('Delete failed');
                })
                .then(() => {
                    uploadResult.innerText = 'File deleted.';
                    return loadSecureFiles();
                })
                .catch(err => {
                    uploadResult.innerText = 'Delete failed: ' + (err.message || err);
                });
        }
    });
});

function register() {
    const email = document.getElementById("email") && document.getElementById("email").value
    const username = document.getElementById("userid") && document.getElementById("userid").value
    const password = document.getElementById("password") && document.getElementById("password").value
    const confirmed_password = document.getElementById("confirm") && document.getElementById("confirm").value

    if (!email || !username || !password) {
        alert('请填写所有必填项');
        return;
    }
    if (password !== confirmed_password) {
        alert("两次密码输入不一致，请重新输入！");
        return;
    }

    const data = { email, username, password, confirmed_password }

    fetch(API_BASE + "/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data)
    })
        .then(res => res.json())
        .then(result => {
            document.getElementById("regResult").innerText = JSON.stringify(result);
            if (isSuccess(result)) {
                // registration success
                window.location.href = "/register_result"
            } else {
                alert(result.message || '注册失败')
            }
        })
        .catch(err => {
            console.error(err)
            alert('注册请求失败')
        })

}

function login() {
    console.log('login() function called');

    // 从表单获取数据
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;
    const resultElement = document.getElementById("loginResult");

    console.log('Email:', email);
    console.log('Password:', password);

    if (!email && !password) {
        alert('请输入邮箱和密码');
        return false;
    } else if(!email) {
        alert('请输入邮箱');
        return false;
    } else if (!password) {
        alert('请输入密码');
        return false;
    }

    // 构建请求数据
    const requestData = {
        email: email,
        password: password
    };

    console.log('Sending request to:', API_BASE + "/auth/login");
    console.log('Request data:', requestData);

    // 发送POST请求
    fetch(API_BASE + "/auth/login", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
        body: JSON.stringify(requestData),
        
    })
    
        .then(response => {
            console.log('Response status:', response.status);
            console.log('Response headers:', response.headers);

            // 首先检查HTTP状态码
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // 尝试解析JSON
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return response.json();
            } else {
                // 如果不是JSON，获取文本
                return response.text().then(text => {
                    console.log('Non-JSON response:', text);
                    throw new Error('Server returned non-JSON response');
                });
            }
        })
        .then(result => {
            console.log('Login response:', result);

            if (isSuccess(result)) {
                // 登录成功，存储token（如果有的话）
                if (result.data && result.data.token) {
                    localStorage.setItem('authToken', result.data.token);
                    localStorage.setItem('userEmail', email);
                    document.cookie = "token=" + encodeURIComponent(result.data.token) + "; path=/; SameSite=Lax";
                    
                }
                // 跳转到首页
                console.log('Redirecting to /index');
                window.location.href = "/index";
            } else {
                // 显示错误信息
                const errorMsg = result.message || result.msg || '登录失败';
                if (resultElement) {
                    resultElement.textContent = errorMsg;
                }
                alert(errorMsg);
            }
        })
        .catch(err => {
            console.error('Login error:', err);

            // 更详细的错误信息
            let errorMsg = '登录请求失败';

            if (err.message.includes('Failed to fetch')) {
                errorMsg = '无法连接到服务器，请检查服务器是否运行';
            } else if (err.message.includes('non-JSON')) {
                errorMsg = '服务器返回了非JSON响应，请检查API';
            } else {
                errorMsg = err.message;
            }

            if (resultElement) {
                resultElement.textContent = errorMsg;
            }
            alert(errorMsg);
        });

    return false; // 阻止表单提交
}

function clearLoginForm() {
    const emailInput = document.getElementById('email');
    const passwordInput = document.getElementById('password');
    
    if (emailInput) {
        emailInput.value = '';
        // 设置空值后再设置一次以确保清空
        emailInput.setAttribute('value', '');
    }
    
    if (passwordInput) {
        passwordInput.value = '';
        // 设置空值后再设置一次以确保清空
        passwordInput.setAttribute('value', '');
    }
    
    // 移除登出标志
    localStorage.removeItem('justLoggedOut');
}

// 页面加载完成后绑定事件
document.addEventListener('DOMContentLoaded', function () {
     // 检查是否是从登出跳转过来的
    const justLoggedOut = localStorage.getItem('justLoggedOut');
    if (justLoggedOut === 'true') {
        clearLoginForm();
    }
    
    // 额外的：如果用户已经登录，直接跳转到首页
    const logout_token = localStorage.getItem('authToken');
    if (logout_token && window.location.pathname === '/login') {
        window.location.href = '/index';
    }

    console.log('DOM loaded, initializing login form');

    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        console.log('Found login form');
        loginForm.addEventListener('submit', function (e) {
            console.log('Form submit event triggered');
            e.preventDefault(); // 阻止默认表单提交
            login();
        });
    } else {
        console.error('Login form not found! Check the HTML.');
    }

    // 为按钮添加点击事件作为备用（避免和内联onclick重复触发）
    const loginButton = document.querySelector('.btn-login');
    if (loginButton && !loginButton.hasAttribute('onclick')) {
        console.log('Found login button');
        loginButton.addEventListener('click', function (e) {
            console.log('Button click event triggered');
            e.preventDefault();
            login();
        });
    }

    // 检查用户是否已登录
    const token = localStorage.getItem('authToken');
    if (token) {
        console.log('User is already logged in with token');
        // 可以选择自动跳转或显示已登录状态
        // window.location.href = "/index";
    }
});

document.addEventListener('DOMContentLoaded', function () {
    const changeForm = document.getElementById('changePasswordForm');
    if (!changeForm) return;
    changeForm.addEventListener('submit', function (e) {
        e.preventDefault();
        changePassword();
    });
});
