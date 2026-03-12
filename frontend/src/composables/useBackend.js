import { ref, reactive } from 'vue'

const Backend = window.go?.main?.App || window.go?.openOMT?.App
const isDesktop = !!Backend

export function useBackend() {
  const connected = ref(false)
  const protocol = ref('')
  const currentPath = ref('/')
  const fileList = ref([])
  const savedConnections = ref([])

  async function fetchStatus() {
    if (!isDesktop) {
      connected.value = false
      protocol.value = ''
      return
    }
    try {
      const status = await Backend.Status()
      connected.value = status.connected
      protocol.value = status.protocol || ''
    } catch {
      connected.value = false
      protocol.value = ''
    }
  }

  async function connect(config) {
    if (!isDesktop) throw new Error('请在桌面程序中运行')
    await Backend.Connect(config)
    await fetchStatus()
    currentPath.value = '/'
  }

  async function connectSaved(id) {
    if (!isDesktop) throw new Error('请在桌面程序中运行')
    await Backend.ConnectSaved(id)
    await fetchStatus()
    currentPath.value = '/'
  }

  async function disconnect() {
    if (!isDesktop) return
    await Backend.Disconnect()
    connected.value = false
    protocol.value = ''
  }

  async function loadList(path) {
    if (!isDesktop) return
    const targetPath = path !== undefined ? path : currentPath.value
    const res = await Backend.List(targetPath)
    if (res.error) throw new Error(res.error)
    currentPath.value = targetPath
    fileList.value = res.list || []
  }

  async function loadSavedConnections() {
    if (!isDesktop || !Backend.GetSavedConnections) {
      savedConnections.value = []
      return
    }
    try {
      const list = await Backend.GetSavedConnections()
      savedConnections.value = Array.isArray(list) ? list : []
    } catch {
      savedConnections.value = []
    }
  }

  async function saveConnection(id, name, host, port, username, password, proto) {
    if (!isDesktop) throw new Error('请在桌面程序中运行')
    await Backend.SaveConnection(id, name, host, port, username, password, proto)
    await loadSavedConnections()
  }

  async function deleteConnection(id) {
    if (!isDesktop) return
    await Backend.DeleteConnection(id)
    await loadSavedConnections()
  }

  async function renameConnection(id, name) {
    if (!isDesktop) return
    await Backend.RenameConnection(id, name)
    await loadSavedConnections()
  }

  async function createDir(path) {
    if (!isDesktop) return
    await Backend.CreateDir(path)
  }

  async function createFile(path) {
    if (!isDesktop) return
    await Backend.CreateFile(path)
  }

  async function deleteFile(path) {
    if (!isDesktop) return
    await Backend.DeleteFile(path)
  }

  async function deleteDirRecursive(path) {
    if (!isDesktop) return
    await Backend.DeleteDirRecursive(path)
  }

  async function download(remotePath) {
    if (!isDesktop) return
    const name = remotePath.split('/').filter(Boolean).pop() || 'download'
    const localPath = await Backend.SaveFileDialog(name)
    if (!localPath) return null
    await Backend.DownloadToPath(remotePath, localPath)
    return localPath
  }

  async function uploadFiles() {
    if (!isDesktop) return []
    const paths = await Backend.OpenMultipleFilesDialog()
    if (!paths || paths.length === 0) return []
    const remoteDir = currentPath.value === '.' ? '' : currentPath.value.replace(/\/$/, '') || ''
    const results = []
    for (const localPath of paths) {
      try {
        await Backend.UploadFromPath(remoteDir, localPath)
        results.push({ path: localPath, success: true })
      } catch (e) {
        results.push({ path: localPath, success: false, error: e.message })
      }
    }
    return results
  }

  async function uploadBase64(remotePath, base64Data) {
    if (!isDesktop) return
    await Backend.UploadBase64(remotePath, base64Data)
  }

  async function executeCommand(cmd) {
    if (!isDesktop) return { output: '', error: '请在桌面程序中运行' }
    return await Backend.ExecuteCommand(cmd)
  }

  async function executeLocalCommand(cmd) {
    if (!isDesktop) return { output: '', error: '请在桌面程序中运行' }
    return await Backend.ExecuteLocalCommand(cmd)
  }

  return {
    isDesktop,
    connected,
    protocol,
    currentPath,
    fileList,
    savedConnections,
    fetchStatus,
    connect,
    connectSaved,
    disconnect,
    loadList,
    loadSavedConnections,
    saveConnection,
    deleteConnection,
    renameConnection,
    createDir,
    createFile,
    deleteFile,
    deleteDirRecursive,
    download,
    uploadFiles,
    uploadBase64,
    executeCommand,
    executeLocalCommand,
  }
}

export const backend = useBackend()
