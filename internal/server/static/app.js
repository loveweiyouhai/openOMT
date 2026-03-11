(function () {
  const API = '/api';
  let currentPath = '/';

  const el = {
    formConnect: document.getElementById('formConnect'),
    protocol: document.getElementById('protocol'),
    host: document.getElementById('host'),
    port: document.getElementById('port'),
    username: document.getElementById('username'),
    password: document.getElementById('password'),
    btnConnect: document.getElementById('btnConnect'),
    btnDisconnect: document.getElementById('btnDisconnect'),
    statusBadge: document.getElementById('statusBadge'),
    currentPath: document.getElementById('currentPath'),
    pathDisplay: document.getElementById('pathDisplay'),
    btnGoUp: document.getElementById('btnGoUp'),
    pathHistory: document.querySelector('.path-history'),
    btnRefresh: document.getElementById('btnRefresh'),
    fileTableBody: document.getElementById('fileTableBody'),
    uploadZone: document.getElementById('uploadZone'),
    fileInput: document.getElementById('fileInput'),
    btnSelectFile: document.getElementById('btnSelectFile'),
    message: document.getElementById('message'),
  };

  function setPortPlaceholder() {
    el.port.placeholder = el.protocol.value === 'sftp' ? '22' : '21';
  }

  function showMessage(text, type) {
    el.message.textContent = text;
    el.message.className = 'message ' + (type || '');
    el.message.hidden = false;
    setTimeout(() => { el.message.hidden = true; }, 4000);
  }

  function setConnected(connected) {
    el.statusBadge.textContent = connected ? '已连接' : '未连接';
    el.statusBadge.className = 'badge ' + (connected ? 'badge-online' : 'badge-offline');
    el.btnDisconnect.disabled = !connected;
    el.btnRefresh.disabled = !connected;
    if (!connected) {
      el.fileTableBody.innerHTML = '<tr class="empty-row"><td colspan="4">请先连接服务器并选择路径</td></tr>';
    }
  }

  async function fetchStatus() {
    try {
      const r = await fetch(API + '/status');
      const d = await r.json();
      setConnected(d.connected);
      return d;
    } catch (e) {
      setConnected(false);
      return { connected: false };
    }
  }

  async function loadList(remotePath) {
    const path = remotePath !== undefined ? remotePath : currentPath;
    try {
      const r = await fetch(API + '/list', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ remotePath: path }),
      });
      const d = await r.json();
      if (!d.ok) {
        showMessage(d.error || '列表加载失败', 'error');
        return;
      }
      currentPath = path;
      el.currentPath.value = path || '/';
      el.pathDisplay.textContent = path || '/';
      renderTable(d.list || []);
    } catch (e) {
      showMessage('网络错误: ' + e.message, 'error');
    }
  }

  function formatSize(bytes) {
    if (bytes === 0 || undefined) return '-';
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    let n = bytes;
    while (n >= 1024 && i < units.length - 1) {
      n /= 1024;
      i++;
    }
    return n.toFixed(i === 0 ? 0 : 1) + ' ' + units[i];
  }

  function renderTable(list) {
    if (!list || list.length === 0) {
      el.fileTableBody.innerHTML = '<tr class="empty-row"><td colspan="4">此目录为空</td></tr>';
      return;
    }
    const rows = list.map((item) => {
      const isDir = item.isDir;
      const rowClass = isDir ? 'row-dir' : 'row-file';
      const size = isDir ? '-' : formatSize(item.size);
      let actions = '';
      if (isDir) {
        actions = '<button type="button" class="btn btn-ghost btn-sm btn-open-dir" data-path="' + escapeAttr(item.path) + '">进入</button>';
      } else {
        actions = '<button type="button" class="btn btn-ghost btn-sm btn-download" data-path="' + escapeAttr(item.path) + '">下载</button>';
      }
      return (
        '<tr class="' + rowClass + '">' +
        '<td class="col-name">' + escapeHtml(item.name) + '</td>' +
        '<td class="col-size">' + size + '</td>' +
        '<td class="col-time">' + escapeHtml(item.modTime || '-') + '</td>' +
        '<td class="col-actions">' + actions + '</td>' +
        '</tr>'
      );
    }).join('');
    el.fileTableBody.innerHTML = rows;

    el.fileTableBody.querySelectorAll('.btn-open-dir').forEach((btn) => {
      btn.addEventListener('click', () => loadList(btn.getAttribute('data-path')));
    });
    el.fileTableBody.querySelectorAll('.btn-download').forEach((btn) => {
      btn.addEventListener('click', () => download(btn.getAttribute('data-path')));
    });
  }

  function escapeHtml(s) {
    const div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
  }

  function escapeAttr(s) {
    return escapeHtml(s).replace(/"/g, '&quot;');
  }

  function download(remotePath) {
    window.location.href = API + '/download?path=' + encodeURIComponent(remotePath);
    showMessage('开始下载: ' + remotePath, 'success');
  }

  el.formConnect.addEventListener('submit', async (e) => {
    e.preventDefault();
    const port = el.port.value ? parseInt(el.port.value, 10) : (el.protocol.value === 'sftp' ? 22 : 21);
    el.btnConnect.disabled = true;
    try {
      const r = await fetch(API + '/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          protocol: el.protocol.value,
          host: el.host.value.trim(),
          port: port,
          username: el.username.value.trim(),
          password: el.password.value,
        }),
      });
      const d = await r.json();
      if (d.ok) {
        setConnected(true);
        currentPath = '/';
        loadList('.');
        showMessage('连接成功', 'success');
      } else {
        showMessage(d.error || '连接失败', 'error');
      }
    } catch (err) {
      showMessage('请求失败: ' + err.message, 'error');
    } finally {
      el.btnConnect.disabled = false;
    }
  });

  el.btnDisconnect.addEventListener('click', async () => {
    try {
      await fetch(API + '/disconnect', { method: 'POST' });
      setConnected(false);
      showMessage('已断开连接', 'success');
    } catch (e) {
      showMessage(e.message, 'error');
    }
  });

  el.protocol.addEventListener('change', setPortPlaceholder);
  setPortPlaceholder();

  el.btnGoUp.addEventListener('click', () => {
    if (!currentPath || currentPath === '/' || currentPath === '.') return;
    const parts = currentPath.replace(/\/$/, '').split('/').filter(Boolean);
    parts.pop();
    const parent = parts.length ? '/' + parts.join('/') : '/';
    loadList(parent);
  });

  el.pathHistory.querySelectorAll('[data-path]').forEach((btn) => {
    btn.addEventListener('click', () => loadList(btn.getAttribute('data-path')));
  });

  el.btnRefresh.addEventListener('click', () => loadList());

  el.uploadZone.addEventListener('click', (e) => {
    if (e.target.id !== 'btnSelectFile') return;
    el.fileInput.click();
  });

  el.fileInput.addEventListener('change', () => {
    const files = el.fileInput.files;
    if (!files.length) return;
    uploadFiles(files);
    el.fileInput.value = '';
  });

  el.uploadZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    el.uploadZone.classList.add('dragover');
  });

  el.uploadZone.addEventListener('dragleave', () => {
    el.uploadZone.classList.remove('dragover');
  });

  el.uploadZone.addEventListener('drop', (e) => {
    e.preventDefault();
    el.uploadZone.classList.remove('dragover');
    const files = e.dataTransfer.files;
    if (files.length) uploadFiles(files);
  });

  async function uploadFiles(files) {
    const status = await fetchStatus();
    if (!status.connected) {
      showMessage('请先连接服务器', 'error');
      return;
    }
    const basePath = currentPath === '.' ? '' : (currentPath.replace(/\/$/, '') || '');
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const fd = new FormData();
      fd.append('file', file);
      fd.append('remotePath', basePath ? basePath + '/' : '');
      try {
        const r = await fetch(API + '/upload', { method: 'POST', body: fd });
        const d = await r.json();
        if (d.ok) {
          showMessage('已上传: ' + file.name, 'success');
          loadList();
        } else {
          showMessage(d.error || '上传失败', 'error');
        }
      } catch (e) {
        showMessage('上传失败: ' + e.message, 'error');
      }
    }
  }

  fetchStatus();
})();
