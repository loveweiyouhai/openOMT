package main

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"openOMT/internal/conn"
	"openOMT/internal/store"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 供前端调用的桌面应用后端
type App struct {
	ctx   context.Context
	pool  *conn.Pool
	store *store.Store
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{pool: conn.GetPool()}
}

// Startup 在窗口就绪后由 Wails 调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	s, err := store.New("")
	if err != nil {
		return
	}
	a.store = s
}

func (a *App) Shutdown(ctx context.Context) {
	if a.store != nil {
		a.store.Close()
	}
	if a.pool != nil {
		a.pool.CloseAll()
	}
}

// ConnectResult 连接结果
type ConnectResult struct {
	ConnId string `json:"connId"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Error  string `json:"error,omitempty"`
}

// Connect 连接服务器，返回连接 ID
func (a *App) Connect(req conn.ConnectRequest) ConnectResult {
	connId, err := a.pool.Connect(req)
	if err != nil {
		return ConnectResult{Error: err.Error()}
	}
	return ConnectResult{
		ConnId: connId,
		Name:   req.Host,
		Host:   req.Host,
		Port:   req.Port,
	}
}

// SavedConnectionItem 供前端绑定的连接项
type SavedConnectionItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Protocol string `json:"protocol"`
}

// GetSavedConnections 返回已保存的连接列表
func (a *App) GetSavedConnections() ([]SavedConnectionItem, error) {
	if a.store == nil {
		return []SavedConnectionItem{}, nil
	}
	list, err := a.store.List()
	if err != nil || list == nil {
		return []SavedConnectionItem{}, nil
	}
	out := make([]SavedConnectionItem, len(list))
	for i, c := range list {
		out[i] = SavedConnectionItem{
			ID: c.ID, Name: c.Name, Host: c.Host, Port: c.Port, Username: c.Username, Protocol: c.Protocol,
		}
	}
	return out, nil
}

// ConnectSaved 使用已保存的连接 ID 连接，返回新的连接 ID
func (a *App) ConnectSaved(savedId string) ConnectResult {
	if a.store == nil {
		return ConnectResult{Error: "存储不可用"}
	}
	req, ok := a.store.GetByID(savedId)
	if !ok {
		return ConnectResult{Error: "连接不存在"}
	}
	saved, _ := a.store.List()
	var name string
	for _, s := range saved {
		if s.ID == savedId {
			name = s.Name
			if name == "" {
				name = s.Host
			}
			break
		}
	}

	connId, err := a.pool.Connect(req)
	if err != nil {
		return ConnectResult{Error: err.Error()}
	}
	return ConnectResult{
		ConnId: connId,
		Name:   name,
		Host:   req.Host,
		Port:   req.Port,
	}
}

// SaveConnection 保存连接
func (a *App) SaveConnection(id, name, host string, port int, username, password, protocol string) (string, error) {
	if a.store == nil {
		return "", errors.New("存储不可用")
	}
	return a.store.Save(id, name, host, port, username, password, protocol)
}

// DeleteConnection 删除已保存的连接
func (a *App) DeleteConnection(id string) error {
	if a.store == nil {
		return errors.New("存储不可用")
	}
	return a.store.Delete(id)
}

// RenameConnection 重命名连接
func (a *App) RenameConnection(id, name string) error {
	if a.store == nil {
		return errors.New("存储不可用")
	}
	return a.store.Rename(id, name)
}

// Disconnect 断开指定连接
func (a *App) Disconnect(connId string) error {
	return a.pool.Disconnect(connId)
}

// ListActiveConnections 列出所有活跃连接
func (a *App) ListActiveConnections() []conn.ConnectionInfo {
	return a.pool.List()
}

// StatusResult 连接状态
type StatusResult struct {
	Connected bool   `json:"connected"`
	Protocol  string `json:"protocol"`
}

// Status 返回指定连接状态
func (a *App) Status(connId string) StatusResult {
	return StatusResult{
		Connected: a.pool.IsConnected(connId),
		Protocol:  string(a.pool.GetProtocol(connId)),
	}
}

// ListResult 目录列表结果
type ListResult struct {
	List  []conn.FileInfo `json:"list"`
	Error string          `json:"error,omitempty"`
}

// List 列出远程目录
func (a *App) List(connId, remotePath string) ListResult {
	list, err := a.pool.ListDir(connId, remotePath)
	if err != nil {
		return ListResult{Error: err.Error()}
	}
	return ListResult{List: list}
}

// SaveFileDialog 打开"另存为"对话框
func (a *App) SaveFileDialog(defaultName string) (string, error) {
	return wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "保存文件",
	})
}

// DownloadToPath 下载文件到本地
func (a *App) DownloadToPath(connId, remotePath, localPath string) error {
	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return a.pool.Download(connId, remotePath, f)
}

// OpenMultipleFilesDialog 打开多选文件对话框
func (a *App) OpenMultipleFilesDialog() ([]string, error) {
	return wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择要上传的文件",
	})
}

// UploadFromPath 从本地路径上传
func (a *App) UploadFromPath(connId, remoteDir, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	remotePath := remoteDir
	if remotePath != "" && remotePath != "." {
		remotePath = remotePath + "/"
	}
	remotePath = remotePath + filepath.Base(localPath)
	return a.pool.Upload(connId, remotePath, f, info.Size())
}

// UploadBase64 上传 base64 编码的文件
func (a *App) UploadBase64(connId, remotePath, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}
	return a.pool.Upload(connId, remotePath, strings.NewReader(string(data)), int64(len(data)))
}

// ExecuteCommandResult 命令执行结果
type ExecuteCommandResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// ExecuteCommand 在远程执行命令
func (a *App) ExecuteCommand(connId, cmd string) ExecuteCommandResult {
	out, err := a.pool.ExecuteCommand(connId, cmd)
	if err != nil {
		errStr := err.Error()
		if out != "" {
			errStr = out + "\n---\n" + errStr
		}
		return ExecuteCommandResult{Output: out, Error: errStr}
	}
	return ExecuteCommandResult{Output: out}
}

// CreateDir 创建目录
func (a *App) CreateDir(connId, remotePath string) error {
	return a.pool.MkdirAll(connId, remotePath)
}

// CreateFile 创建空文件
func (a *App) CreateFile(connId, remotePath string) error {
	return a.pool.CreateEmptyFile(connId, remotePath)
}

// DeleteFile 删除文件
func (a *App) DeleteFile(connId, remotePath string) error {
	return a.pool.DeleteFile(connId, remotePath)
}

// DeleteDirRecursive 递归删除目录
func (a *App) DeleteDirRecursive(connId, remotePath string) error {
	return a.pool.DeleteDirRecursive(connId, remotePath)
}

// ExecuteLocalCommand 在本机执行命令
func (a *App) ExecuteLocalCommand(cmd string) ExecuteCommandResult {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd.exe", "/c", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}
	out, err := c.CombinedOutput()
	output := strings.TrimSuffix(string(out), "\n")
	if err != nil {
		errStr := err.Error()
		if output != "" {
			errStr = output + "\n---\n" + errStr
		}
		return ExecuteCommandResult{Output: output, Error: errStr}
	}
	return ExecuteCommandResult{Output: output}
}

// ShellResult shell 启动结果
type ShellResult struct {
	ShellId string `json:"shellId"`
	Error   string `json:"error,omitempty"`
}

// StartShell 启动交互式 shell 会话
func (a *App) StartShell(connId string) ShellResult {
	shellId, err := a.pool.StartShell(connId)
	if err != nil {
		return ShellResult{Error: err.Error()}
	}
	// 启动读取协程
	go a.readShellOutput(connId, shellId)
	return ShellResult{ShellId: shellId}
}

// readShellOutput 读取 shell 输出并发送到前端
func (a *App) readShellOutput(connId, shellId string) {
	buf := make([]byte, 4096)
	eventKey := "shell-output:" + connId + ":" + shellId
	closeKey := "shell-closed:" + connId + ":" + shellId
	for {
		n, err := a.pool.ReadShell(connId, shellId, buf)
		if err != nil {
			wailsruntime.EventsEmit(a.ctx, closeKey)
			return
		}
		if n > 0 {
			wailsruntime.EventsEmit(a.ctx, eventKey, string(buf[:n]))
		}
	}
}

// WriteShell 向 shell 写入数据
func (a *App) WriteShell(connId, shellId, data string) error {
	return a.pool.WriteShell(connId, shellId, data)
}

// ResizeShell 调整终端大小
func (a *App) ResizeShell(connId, shellId string, rows, cols int) error {
	return a.pool.ResizeShell(connId, shellId, rows, cols)
}

// CloseShell 关闭 shell 会话
func (a *App) CloseShell(connId, shellId string) error {
	return a.pool.CloseShell(connId, shellId)
}

// ListShells 列出连接的所有终端
func (a *App) ListShells(connId string) []string {
	return a.pool.ListShells(connId)
}
