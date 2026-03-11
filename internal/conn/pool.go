package conn

import (
	"sync"

	"github.com/google/uuid"
)

// Connection 单个连接实例
type Connection struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Protocol Protocol
	client   interface{ Close() error }
	shells   map[string]*ShellSession // 多个终端会话
}

// ConnectionInfo 连接信息（供前端显示）
type ConnectionInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// Pool 连接池，管理多个并发连接
type Pool struct {
	mu    sync.RWMutex
	conns map[string]*Connection
}

var defaultPool = &Pool{
	conns: make(map[string]*Connection),
}

// GetPool 获取全局连接池
func GetPool() *Pool {
	return defaultPool
}

// Connect 建立新连接，返回连接 ID
func (p *Pool) Connect(req ConnectRequest) (string, error) {
	connId := uuid.New().String()

	var client interface{ Close() error }
	var err error

	switch req.Protocol {
	case ProtocolSFTP:
		var c *SFTPClient
		c, err = NewSFTPClient(req)
		if err == nil {
			client = c
		}
	case ProtocolFTP:
		var c *FTPClient
		c, err = NewFTPClient(req)
		if err == nil {
			client = c
		}
	default:
		return "", ErrUnsupportedProtocol
	}

	if err != nil {
		return "", err
	}

	conn := &Connection{
		ID:       connId,
		Name:     req.Host,
		Host:     req.Host,
		Port:     req.Port,
		Protocol: req.Protocol,
		client:   client,
	}

	p.mu.Lock()
	p.conns[connId] = conn
	p.mu.Unlock()

	return connId, nil
}

// ConnectWithID 使用指定 ID 建立连接（用于恢复已保存的连接）
func (p *Pool) ConnectWithID(connId string, req ConnectRequest, name string) error {
	var client interface{ Close() error }
	var err error

	switch req.Protocol {
	case ProtocolSFTP:
		var c *SFTPClient
		c, err = NewSFTPClient(req)
		if err == nil {
			client = c
		}
	case ProtocolFTP:
		var c *FTPClient
		c, err = NewFTPClient(req)
		if err == nil {
			client = c
		}
	default:
		return ErrUnsupportedProtocol
	}

	if err != nil {
		return err
	}

	displayName := name
	if displayName == "" {
		displayName = req.Host
	}

	conn := &Connection{
		ID:       connId,
		Name:     displayName,
		Host:     req.Host,
		Port:     req.Port,
		Protocol: req.Protocol,
		client:   client,
	}

	p.mu.Lock()
	p.conns[connId] = conn
	p.mu.Unlock()

	return nil
}

// Disconnect 断开指定连接
func (p *Pool) Disconnect(connId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, ok := p.conns[connId]
	if !ok {
		return nil
	}

	// 关闭所有 shell 会话
	if conn.shells != nil {
		for _, shell := range conn.shells {
			shell.Close()
		}
		conn.shells = nil
	}

	var err error
	if conn.client != nil {
		err = conn.client.Close()
	}
	delete(p.conns, connId)
	return err
}

// Get 获取指定连接
func (p *Pool) Get(connId string) *Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.conns[connId]
}

// List 列出所有活跃连接
func (p *Pool) List() []ConnectionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	list := make([]ConnectionInfo, 0, len(p.conns))
	for _, conn := range p.conns {
		list = append(list, ConnectionInfo{
			ID:       conn.ID,
			Name:     conn.Name,
			Host:     conn.Host,
			Port:     conn.Port,
			Protocol: string(conn.Protocol),
		})
	}
	return list
}

// IsConnected 检查指定连接是否存在
func (p *Pool) IsConnected(connId string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.conns[connId]
	return ok
}

// CloseAll 关闭所有连接
func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, conn := range p.conns {
		if conn.client != nil {
			conn.client.Close()
		}
		delete(p.conns, id)
	}
}

// ListDir 列出远程目录
func (p *Pool) ListDir(connId, remotePath string) ([]FileInfo, error) {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return nil, ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.List(remotePath)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.List(remotePath)
	}
	return nil, ErrNotConnected
}

// Download 下载文件
func (p *Pool) Download(connId, remotePath string, w interface{ Write(p []byte) (n int, err error) }) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.Download(remotePath, w)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.Download(remotePath, w)
	}
	return ErrNotConnected
}

// Upload 上传文件
func (p *Pool) Upload(connId, remotePath string, r interface{ Read(p []byte) (n int, err error) }, size int64) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.Upload(remotePath, r, size)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.Upload(remotePath, r, size)
	}
	return ErrNotConnected
}

// MkdirAll 创建目录
func (p *Pool) MkdirAll(connId, remotePath string) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.MkdirAll(remotePath)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.MkdirAll(remotePath)
	}
	return ErrNotConnected
}

// CreateEmptyFile 创建空文件
func (p *Pool) CreateEmptyFile(connId, remotePath string) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.CreateEmptyFile(remotePath)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.CreateEmptyFile(remotePath)
	}
	return ErrNotConnected
}

// DeleteFile 删除文件
func (p *Pool) DeleteFile(connId, remotePath string) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.DeleteFile(remotePath)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.DeleteFile(remotePath)
	}
	return ErrNotConnected
}

// DeleteDirRecursive 递归删除目录
func (p *Pool) DeleteDirRecursive(connId, remotePath string) error {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.DeleteDirRecursive(remotePath)
	}
	if c, ok := conn.client.(*FTPClient); ok {
		return c.DeleteDirRecursive(remotePath)
	}
	return ErrNotConnected
}

// ExecuteCommand 执行远程命令（仅 SFTP）
func (p *Pool) ExecuteCommand(connId, cmd string) (string, error) {
	conn := p.Get(connId)
	if conn == nil || conn.client == nil {
		return "", ErrNotConnected
	}

	if c, ok := conn.client.(*SFTPClient); ok {
		return c.ExecuteCommand(cmd)
	}
	return "", ErrTerminalNotSupported
}

// GetProtocol 获取连接协议
func (p *Pool) GetProtocol(connId string) Protocol {
	conn := p.Get(connId)
	if conn == nil {
		return ""
	}
	return conn.Protocol
}

// StartShell 启动交互式 shell 会话，返回 shellId
func (p *Pool) StartShell(connId string) (string, error) {
	p.mu.Lock()
	conn := p.conns[connId]
	p.mu.Unlock()

	if conn == nil || conn.client == nil {
		return "", ErrNotConnected
	}

	sftpClient, ok := conn.client.(*SFTPClient)
	if !ok {
		return "", ErrTerminalNotSupported
	}

	shell, err := NewShellSession(sftpClient.ssh)
	if err != nil {
		return "", err
	}

	shellId := uuid.New().String()

	p.mu.Lock()
	if conn.shells == nil {
		conn.shells = make(map[string]*ShellSession)
	}
	conn.shells[shellId] = shell
	p.mu.Unlock()

	return shellId, nil
}

// WriteShell 向 shell 写入数据
func (p *Pool) WriteShell(connId, shellId, data string) error {
	conn := p.Get(connId)
	if conn == nil {
		return ErrNotConnected
	}
	if conn.shells == nil {
		return ErrTerminalNotSupported
	}
	shell := conn.shells[shellId]
	if shell == nil {
		return ErrTerminalNotSupported
	}
	return shell.Write(data)
}

// ReadShell 从 shell 读取数据
func (p *Pool) ReadShell(connId, shellId string, buf []byte) (int, error) {
	conn := p.Get(connId)
	if conn == nil {
		return 0, ErrNotConnected
	}
	if conn.shells == nil {
		return 0, ErrTerminalNotSupported
	}
	shell := conn.shells[shellId]
	if shell == nil {
		return 0, ErrTerminalNotSupported
	}
	return shell.ReadOutput(buf)
}

// ResizeShell 调整 shell 终端大小
func (p *Pool) ResizeShell(connId, shellId string, rows, cols int) error {
	conn := p.Get(connId)
	if conn == nil {
		return ErrNotConnected
	}
	if conn.shells == nil {
		return ErrTerminalNotSupported
	}
	shell := conn.shells[shellId]
	if shell == nil {
		return ErrTerminalNotSupported
	}
	return shell.Resize(rows, cols)
}

// CloseShell 关闭指定 shell 会话
func (p *Pool) CloseShell(connId, shellId string) error {
	p.mu.Lock()
	conn := p.conns[connId]
	p.mu.Unlock()

	if conn == nil || conn.shells == nil {
		return nil
	}
	shell := conn.shells[shellId]
	if shell != nil {
		err := shell.Close()
		delete(conn.shells, shellId)
		return err
	}
	return nil
}

// CloseAllShells 关闭连接的所有 shell
func (p *Pool) CloseAllShells(connId string) {
	p.mu.Lock()
	conn := p.conns[connId]
	p.mu.Unlock()

	if conn == nil || conn.shells == nil {
		return
	}
	for id, shell := range conn.shells {
		shell.Close()
		delete(conn.shells, id)
	}
}

// ListShells 列出连接的所有 shell
func (p *Pool) ListShells(connId string) []string {
	conn := p.Get(connId)
	if conn == nil || conn.shells == nil {
		return nil
	}
	ids := make([]string, 0, len(conn.shells))
	for id := range conn.shells {
		ids = append(ids, id)
	}
	return ids
}
