package conn

import (
	"sync"
)

// Client 统一文件客户端接口
type Client interface {
	List(remotePath string) ([]FileInfo, error)
	Close() error
}

// Manager 连接管理器（单例，仅维护当前连接）
type Manager struct {
	mu     sync.Mutex
	client interface{ Close() error }
	proto  Protocol
}

var defaultManager = &Manager{}

// GetManager 获取全局连接管理器
func GetManager() *Manager {
	return defaultManager
}

// Connect 建立新连接，会关闭已有连接
func (m *Manager) Connect(req ConnectRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.closeLocked(); err != nil {
		// 忽略关闭错误
	}
	var err error
	switch req.Protocol {
	case ProtocolSFTP:
		var c *SFTPClient
		c, err = NewSFTPClient(req)
		if err == nil {
			m.client = c
			m.proto = ProtocolSFTP
		}
	case ProtocolFTP:
		var c *FTPClient
		c, err = NewFTPClient(req)
		if err == nil {
			m.client = c
			m.proto = ProtocolFTP
		}
	default:
		err = ErrUnsupportedProtocol
	}
	return err
}

// Close 关闭当前连接
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeLocked()
}

func (m *Manager) closeLocked() error {
	if m.client != nil {
		err := m.client.Close()
		m.client = nil
		return err
	}
	return nil
}

// List 列出远程目录（需已连接）
func (m *Manager) List(remotePath string) ([]FileInfo, error) {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return nil, ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.List(remotePath)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.List(remotePath)
	}
	return nil, ErrNotConnected
}

// Download 下载文件（需已连接）
func (m *Manager) Download(remotePath string, w interface{ Write(p []byte) (n int, err error) }) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.Download(remotePath, w)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.Download(remotePath, w)
	}
	return ErrNotConnected
}

// Upload 上传文件（需已连接）
func (m *Manager) Upload(remotePath string, r interface{ Read(p []byte) (n int, err error) }, size int64) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.Upload(remotePath, r, size)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.Upload(remotePath, r, size)
	}
	return ErrNotConnected
}

// IsConnected 是否已连接
func (m *Manager) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client != nil
}

// Protocol 当前协议
func (m *Manager) Protocol() Protocol {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.proto
}

// ExecuteCommand 在远程服务器执行命令（仅 SFTP 支持）
func (m *Manager) ExecuteCommand(cmd string) (output string, err error) {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return "", ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.ExecuteCommand(cmd)
	}
	return "", ErrTerminalNotSupported
}

// MkdirAll 创建远程目录
func (m *Manager) MkdirAll(remotePath string) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.MkdirAll(remotePath)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.MkdirAll(remotePath)
	}
	return ErrNotConnected
}

// CreateEmptyFile 在远程创建空文件
func (m *Manager) CreateEmptyFile(remotePath string) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.CreateEmptyFile(remotePath)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.CreateEmptyFile(remotePath)
	}
	return ErrNotConnected
}

// DeleteFile 在远程删除单个文件
func (m *Manager) DeleteFile(remotePath string) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.DeleteFile(remotePath)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.DeleteFile(remotePath)
	}
	return ErrNotConnected
}

// DeleteDirRecursive 在远程递归删除目录
func (m *Manager) DeleteDirRecursive(remotePath string) error {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	if client == nil {
		return ErrNotConnected
	}
	if c, ok := client.(*SFTPClient); ok {
		return c.DeleteDirRecursive(remotePath)
	}
	if c, ok := client.(*FTPClient); ok {
		return c.DeleteDirRecursive(remotePath)
	}
	return ErrNotConnected
}
