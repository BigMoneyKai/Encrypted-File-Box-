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

function getFileExt(name) {
    const base = String(name || '').toLowerCase();
    const idx = base.lastIndexOf('.');
    if (idx <= 0 || idx === base.length - 1) return '';
    return base.slice(idx + 1);
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
    const filesGrid = document.getElementById('filesGrid');
    if (!filesGrid) return;

    if (!items || items.length === 0) {
        filesGrid.innerHTML = '<div class="text-muted">No files in vault.</div>';
        return;
    }

    filesGrid.innerHTML = items.map(file => {
        const created = file.created_at ? new Date(file.created_at).toLocaleString() : '-';
        const safeName = escapeHTML(file.filename || '-');
        const safeDesc = escapeHTML(file.description || '-');
        const ext = getFileExt(file.filename || '');
        return `
        <div class="file-card" data-id="${file.id}" data-ext="${escapeHTML(ext)}">
            <input type="checkbox" class="select-checkbox" data-id="${file.id}">
            <div class="file-title">${safeName}</div>
            <div class="file-meta">${formatSize(Number(file.size || 0))} • ${escapeHTML(created)}</div>
            <div class="file-meta">${safeDesc}</div>
            <div class="preview" data-preview="${file.id}">
                <div class="text-muted" style="padding:0.6rem;">Preview not loaded</div>
            </div>
            <div class="file-actions">
                <button class="file-action" data-action="preview" data-id="${file.id}">Preview</button>
                <button class="file-action" data-action="download" data-id="${file.id}" data-name="${safeName}">Download</button>
                <button class="file-action update" data-action="update" data-id="${file.id}">Update</button>
                <button class="file-action delete" data-action="delete" data-id="${file.id}">Delete</button>
            </div>
        </div>`;
    }).join('');
}

function loadSecureFiles() {
    const filesGrid = document.getElementById('filesGrid');
    if (!filesGrid) return Promise.resolve();

    const token = getAuthToken();
    const refreshBtn = document.getElementById('refreshFilesBtn');
    if (!token) {
        filesGrid.innerHTML = '<div class="text-muted">Sign in to view protected files.</div>';
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

    filesGrid.innerHTML = '<div class="text-muted">Loading files...</div>';
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
            filesGrid.innerHTML = `<div class="text-muted">Failed to load: ${escapeHTML(err.message || err)}</div>`;
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
    const filesGrid = document.getElementById('filesGrid');
    const uploadQueue = document.getElementById('uploadQueue');
    const globalProgressBar = document.getElementById('globalProgressBar');
    const globalProgressText = document.getElementById('globalProgressText');
    const deleteSelectedBtn = document.getElementById('deleteSelectedBtn');

    if (!uploadBtn || !fileInput || !uploadResult || !filesGrid) return;

    loadSecureFiles();

    const RESUMABLE_THRESHOLD = 10 * 1024 * 1024;
    const CHUNK_SIZE = 2 * 1024 * 1024;
    let queueItems = [];
    const preparedSessions = new Map();
    const makeFileKey = (file) => `${file.name}::${file.size}::${file.lastModified}`;

    const setGlobalProgress = (completed, total) => {
        if (!globalProgressBar || !globalProgressText) return;
        const percent = total > 0 ? Math.round((completed / total) * 100) : 0;
        globalProgressBar.style.width = `${percent}%`;
        globalProgressText.innerText = `[${renderBlocks(percent)}] ${percent}% (${completed} / ${total} files)`;
    };

    const setGlobalPreparing = (total) => {
        if (!globalProgressBar || !globalProgressText) return;
        const percent = total === 1 ? 100 : (total > 0 ? 5 : 0);
        globalProgressBar.style.width = `${percent}%`;
        globalProgressText.innerText = `[${renderBlocks(percent)}] ${percent}% (0 / ${total} files) Ready`;
    };

    const renderBlocks = (percent) => {
        const total = 12;
        const filled = Math.round((percent / 100) * total);
        return `${'█'.repeat(filled)}${'░'.repeat(total - filled)}`;
    };

    const createUploadItem = (file) => {
        if (!uploadQueue) return null;
        const item = document.createElement('div');
        item.className = 'upload-item';
        item.innerHTML = `
            <div class="meta">
                <div>${escapeHTML(file.name)}</div>
                <div>${formatSize(file.size)}</div>
            </div>
            <div class="progress-track"><div class="progress-bar"></div></div>
            <div class="status">Waiting...</div>
        `;
        uploadQueue.appendChild(item);
        return {
            el: item,
            bar: item.querySelector('.progress-bar'),
            status: item.querySelector('.status')
        };
    };

    const setItemProgress = (ui, percent, statusText) => {
        if (!ui) return;
        if (ui.bar) ui.bar.style.width = `${percent}%`;
        if (ui.status && statusText) ui.status.innerText = statusText;
    };

    const setItemPrepared = (ui) => {
        setItemProgress(ui, 5, 'Ready to upload');
    };

    const uploadSingleNormal = (file, description, ui) => {
        return new Promise((resolve, reject) => {
            const token = getAuthToken();
            if (!token) {
                reject(new Error('Session expired. Please log in again.'));
                return;
            }
            const xhr = new XMLHttpRequest();
            xhr.open('POST', API_BASE + '/files/upload');
            xhr.setRequestHeader('Authorization', 'Bearer ' + token);
            xhr.upload.onprogress = (evt) => {
                if (evt.lengthComputable) {
                    const percent = Math.round((evt.loaded / evt.total) * 100);
                    setItemProgress(ui, percent, `Uploading... ${percent}%`);
                }
            };
            xhr.onload = () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    setItemProgress(ui, 100, 'Uploaded');
                    resolve();
                } else {
                    reject(new Error('Upload failed'));
                }
            };
            xhr.onerror = () => reject(new Error('Upload failed'));

            const fd = new FormData();
            fd.append('file', file);
            fd.append('description', description || '');
            xhr.send(fd);
        });
    };

    const initResumable = (file, description) => {
        return fetch(API_BASE + '/files/resumable/init', {
            method: 'POST',
            headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
            body: JSON.stringify({
                filename: file.name,
                total_size: file.size,
                chunk_size: CHUNK_SIZE,
                description: description || ''
            })
        }).then(res => parseJsonSafe(res).then(data => ({ res, data })))
          .then(({ res, data }) => {
              if (!res.ok || !isSuccess(data)) {
                  throw new Error(getErrorMessage(data, 'Init failed'));
              }
              return data.data || data;
          });
    };

    const getResumableStatus = (uploadId) => {
        return fetch(API_BASE + '/files/resumable/' + uploadId, {
            method: 'GET',
            headers: getAuthHeaders()
        }).then(res => parseJsonSafe(res).then(data => ({ res, data })))
          .then(({ res, data }) => {
              if (!res.ok || !isSuccess(data)) {
                  throw new Error(getErrorMessage(data, 'Status failed'));
              }
              return data.data || data;
          });
    };

    const uploadChunk = (uploadId, index, blob, ui, uploadedBytes, totalBytes) => {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            xhr.open('POST', API_BASE + '/files/resumable/' + uploadId + '/chunk');
            const headers = getAuthHeaders();
            Object.keys(headers).forEach(k => xhr.setRequestHeader(k, headers[k]));
            xhr.upload.onprogress = (evt) => {
                if (evt.lengthComputable) {
                    const current = uploadedBytes + evt.loaded;
                    const percent = Math.round((current / totalBytes) * 100);
                    setItemProgress(ui, percent, `Uploading... ${percent}%`);
                }
            };
            xhr.onload = () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    resolve();
                } else {
                    reject(new Error('Chunk upload failed'));
                }
            };
            xhr.onerror = () => reject(new Error('Chunk upload failed'));

            const fd = new FormData();
            fd.append('index', String(index));
            fd.append('chunk', blob, `chunk_${index}`);
            xhr.send(fd);
        });
    };

    const completeResumable = (uploadId) => {
        return fetch(API_BASE + '/files/resumable/' + uploadId + '/complete', {
            method: 'POST',
            headers: getAuthHeaders()
        }).then(res => parseJsonSafe(res).then(data => ({ res, data })))
          .then(({ res, data }) => {
              if (!res.ok || !isSuccess(data)) {
                  throw new Error(getErrorMessage(data, 'Complete failed'));
              }
              return data.data || data;
          });
    };

    const ensurePreparedSession = (file, description, ui) => {
        const key = makeFileKey(file);
        const existing = preparedSessions.get(key);
        if (existing && existing.description === description) {
            if (existing.session) return Promise.resolve(existing.session);
            if (existing.promise) return existing.promise;
        }

        const prep = {
            description,
            session: null,
            promise: null
        };
        prep.promise = initResumable(file, description)
            .then(session => {
                prep.session = session;
                prep.promise = null;
                return session;
            })
            .catch(err => {
                preparedSessions.delete(key);
                throw err;
            });
        preparedSessions.set(key, prep);
        if (ui) setItemProgress(ui, 3, 'Preparing...');
        return prep.promise
            .then(session => {
                if (ui) setItemPrepared(ui);
                return session;
            })
            .catch(() => {
                if (ui) setItemPrepared(ui);
                return null;
            });
    };

    const uploadSingleResumable = async (file, description, ui) => {
        const prepared = await ensurePreparedSession(file, description, ui);
        const session = prepared || await initResumable(file, description);
        const status = await getResumableStatus(session.upload_id);
        const uploadedSet = new Set(status.uploaded || []);
        let uploadedBytes = 0;

        for (let i = 0; i < status.total_chunks; i++) {
            const start = i * status.chunk_size;
            const end = Math.min(start + status.chunk_size, file.size);
            const chunkSize = end - start;
            if (uploadedSet.has(i)) {
                uploadedBytes += chunkSize;
                continue;
            }
            const blob = file.slice(start, end);
            await uploadChunk(status.upload_id, i, blob, ui, uploadedBytes, file.size);
            uploadedBytes += chunkSize;
        }
        setItemProgress(ui, 100, 'Finalizing...');
        await completeResumable(status.upload_id);
        setItemProgress(ui, 100, 'Uploaded');
    };

    const uploadSingle = async (file, description, ui) => {
        if (file.size >= RESUMABLE_THRESHOLD) {
            await uploadSingleResumable(file, description, ui);
        } else {
            await uploadSingleNormal(file, description, ui);
        }
    };

    const rebuildQueue = () => {
        queueItems = [];
        preparedSessions.clear();
        if (uploadQueue) uploadQueue.innerHTML = '';
        const files = fileInput.files ? Array.from(fileInput.files) : [];
        files.forEach(file => {
            const ui = createUploadItem(file);
            setItemPrepared(ui);
            queueItems.push({ file, ui });
        });
        setGlobalPreparing(files.length);
        const description = fileDesc ? fileDesc.value.trim() : '';
        files.forEach((file, idx) => {
            if (file.size >= RESUMABLE_THRESHOLD) {
                const ui = queueItems[idx] ? queueItems[idx].ui : null;
                ensurePreparedSession(file, description, ui).catch(() => {});
            }
        });
    };

    if (fileInput) {
        fileInput.addEventListener('change', rebuildQueue);
    }

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
    uploadBtn.addEventListener('click', async () => {
        uploadResult.innerText = '';
        const files = fileInput.files ? Array.from(fileInput.files) : [];
        if (files.length === 0) {
            uploadResult.innerText = 'Please choose files first.';
            return;
        }
        const token = getAuthToken();
        if (!token) {
            uploadResult.innerText = 'Session expired. Please log in again.';
            redirectToLogin();
            return;
        }

        if (uploadBtn) {
            uploadBtn.disabled = true;
            uploadBtn.innerText = 'Uploading...';
        }

        let completed = 0;
        setGlobalProgress(completed, files.length);
        const description = fileDesc ? fileDesc.value.trim() : '';

        try {
            if (queueItems.length !== files.length) {
                rebuildQueue();
            }
            for (const entry of queueItems) {
                const ui = entry.ui || createUploadItem(entry.file);
                setItemProgress(ui, 0, 'Uploading...');
                await uploadSingle(entry.file, description, ui);
                completed += 1;
                setGlobalProgress(completed, files.length);
            }
            uploadResult.innerText = 'All uploads completed.';
            fileInput.value = '';
            if (fileDesc) fileDesc.value = '';
            await loadSecureFiles();
        } catch (err) {
            uploadResult.innerText = 'Upload failed: ' + (err.message || err);
        } finally {
            if (uploadBtn) {
                uploadBtn.disabled = false;
                uploadBtn.innerText = uploadBtn.dataset.label || 'Upload';
            }
        }
    });

    if (uploadPublicBtn) {
        uploadPublicBtn.dataset.label = 'Upload Publicly';
        uploadPublicBtn.addEventListener('click', () => {
            uploadResult.innerText = 'Public upload is not enabled in this view.';
        });
    }

    if (refreshFilesBtn) {
        refreshFilesBtn.addEventListener('click', () => {
            loadSecureFiles();
        });
    }

    filesGrid.addEventListener('click', (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;
        const action = target.dataset.action;
        const fileId = target.dataset.id;
        if (!action || !fileId) return;

        if (action === 'preview') {
            const card = target.closest('.file-card');
            if (!card) return;
            const previewEl = card.querySelector('.preview');
            if (!previewEl) return;
            const ext = (card.dataset.ext || '').toLowerCase();
            const officeExts = new Set(['doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx']);
            const imageExts = new Set(['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'tif', 'tiff', 'svg', 'heic', 'heif', 'avif']);
            const textExts = new Set(['txt', 'md', 'json', 'log', 'csv']);
            const previewUrl = API_BASE + '/files/preview/' + fileId;

            if (officeExts.has(ext) || imageExts.has(ext) || ext === 'pdf') {
                previewEl.innerHTML = '<div class="text-muted" style="padding:0.6rem;">Opening preview in a new tab...</div>';
                const win = window.open(previewUrl, '_blank', 'noopener');
                if (!win) {
                    previewEl.innerHTML = `<div class="text-muted" style="padding:0.6rem;">Popup blocked. <a href="${previewUrl}" target="_blank" rel="noopener">Open preview</a></div>`;
                }
                return;
            }

            previewEl.innerHTML = '<div class="text-muted" style="padding:0.6rem;">Loading preview...</div>';
            fetch(previewUrl, {
                method: 'GET',
                headers: getAuthHeaders()
            })
                .then(res => {
                    if (!res.ok) throw new Error('Preview failed');
                    const contentType = (res.headers.get('Content-Type') || '').toLowerCase();
                    if (contentType.startsWith('image/') || imageExts.has(ext)) {
                        return res.blob().then(blob => ({ type: 'image', blob }));
                    }
                    if (contentType === 'application/pdf' || ext === 'pdf') {
                        return res.blob().then(blob => ({ type: 'pdf', blob }));
                    }
                    if (contentType.startsWith('text/') || textExts.has(ext) || contentType.includes('json')) {
                        return res.text().then(text => ({ type: 'text', text }));
                    }
                    return res.blob().then(blob => ({ type: 'binary', blob }));
                })
                .then(payload => {
                    if (payload.type === 'image') {
                        const url = URL.createObjectURL(payload.blob);
                        previewEl.innerHTML = `<img src="${url}" alt="preview">`;
                        return;
                    }
                    if (payload.type === 'pdf') {
                        const url = URL.createObjectURL(payload.blob);
                        previewEl.innerHTML = `<iframe class="text-preview" style="width:100%;height:240px;border:0;" src="${url}"></iframe>`;
                        return;
                    }
                    if (payload.type === 'text') {
                        const safe = escapeHTML(payload.text || '').slice(0, 20000);
                        previewEl.innerHTML = `<div class="text-preview">${safe}</div>`;
                        return;
                    }
                    previewEl.innerHTML = `<div class="text-muted" style="padding:0.6rem;">Preview not supported. <a href="${previewUrl}" target="_blank" rel="noopener">Open</a></div>`;
                })
                .catch(err => {
                    previewEl.innerHTML = `<div class="text-muted" style="padding:0.6rem;">${escapeHTML(err.message || err)}</div>`;
                });
            return;
        }

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

    if (deleteSelectedBtn) {
        deleteSelectedBtn.addEventListener('click', () => {
            const checked = filesGrid.querySelectorAll('.select-checkbox:checked');
            const ids = Array.from(checked).map(el => Number(el.dataset.id)).filter(Boolean);
            if (ids.length === 0) {
                uploadResult.innerText = 'Select files to delete.';
                return;
            }
            if (!confirm(`Delete ${ids.length} selected files?`)) return;
            deleteSelectedBtn.disabled = true;
            fetch(API_BASE + '/files/batch', {
                method: 'DELETE',
                headers: getAuthHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify({ ids })
            })
                .then(res => parseJsonSafe(res).then(data => ({ res, data })))
                .then(({ res, data }) => {
                    if (!res.ok || !isSuccess(data)) {
                        throw new Error(getErrorMessage(data, 'Batch delete failed'));
                    }
                    uploadResult.innerText = 'Batch delete completed.';
                    return loadSecureFiles();
                })
                .catch(err => {
                    uploadResult.innerText = 'Batch delete failed: ' + (err.message || err);
                })
                .finally(() => {
                    deleteSelectedBtn.disabled = false;
                });
        });
    }
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
    
    // 额外的：如果本地存在 token，先验证再跳转，避免无效 token 造成循环跳转
    const localToken = localStorage.getItem('authToken');
    if (localToken && window.location.pathname === '/login') {
        fetch(API_BASE + "/user/profile", {
            method: "GET",
            headers: getAuthHeaders({ "Accept": "application/json" })
        })
            .then(res => {
                if (res.ok) {
                    window.location.href = '/index';
                    return;
                }
                clearAuthState();
            })
            .catch(() => {
                clearAuthState();
            });
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
