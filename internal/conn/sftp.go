package conn

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPClient SFTP 连接封装
type SFTPClient struct {
	client *sftp.Client
	ssh    *ssh.Client
}

// NewSFTPClient 创建 SFTP 客户端
func NewSFTPClient(req ConnectRequest) (*SFTPClient, error) {
	password := req.Password
	config := &ssh.ClientConfig{
		User: req.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			// 部分服务器仅支持键盘交互认证，用同一密码应答
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = password
				}
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	port := req.Port
	if port <= 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", req.Host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial: %w", err)
	}
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("sftp new client: %w", err)
	}
	return &SFTPClient{client: client, ssh: sshClient}, nil
}

// List 列出目录
func (c *SFTPClient) List(remotePath string) ([]FileInfo, error) {
	if remotePath == "" {
		remotePath = "."
	}
	entries, err := c.client.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}
	var list []FileInfo
	for _, e := range entries {
		fullPath := path.Join(remotePath, e.Name())
		if remotePath == "." {
			fullPath = e.Name()
		}
		list = append(list, FileInfo{
			Name:    e.Name(),
			Path:    fullPath,
			Size:    e.Size(),
			ModTime: e.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   e.IsDir(),
		})
	}
	return list, nil
}

// Download 下载文件到 Writer
func (c *SFTPClient) Download(remotePath string, w io.Writer) error {
	f, err := c.client.Open(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

// Upload 从 Reader 上传文件
func (c *SFTPClient) Upload(remotePath string, r io.Reader, size int64) error {
	dir := path.Dir(remotePath)
	if dir != "." {
		if err := c.client.MkdirAll(dir); err != nil {
			// 忽略已存在
		}
	}
	f, err := c.client.Create(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// MkdirAll 创建远程目录（含父目录）
func (c *SFTPClient) MkdirAll(remotePath string) error {
	return c.client.MkdirAll(remotePath)
}

// CreateEmptyFile 创建空文件
func (c *SFTPClient) CreateEmptyFile(remotePath string) error {
	f, err := c.client.Create(remotePath)
	if err != nil {
		return err
	}
	return f.Close()
}

// DeleteFile 删除单个文件
func (c *SFTPClient) DeleteFile(remotePath string) error {
	return c.client.Remove(remotePath)
}

// DeleteDirRecursive 递归删除目录及其内容
func (c *SFTPClient) DeleteDirRecursive(remotePath string) error {
	entries, err := c.client.ReadDir(remotePath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		child := path.Join(remotePath, e.Name())
		if e.IsDir() {
			if err := c.DeleteDirRecursive(child); err != nil {
				return err
			}
		} else {
			if err := c.client.Remove(child); err != nil {
				return err
			}
		}
	}
	// 删除空目录本身
	return c.client.RemoveDirectory(remotePath)
}

// ExecuteCommand 在远程服务器上执行 shell 命令，返回标准输出与标准错误合并结果
func (c *SFTPClient) ExecuteCommand(cmd string) (output string, err error) {
	if c.ssh == nil {
		return "", fmt.Errorf("ssh 连接不可用")
	}
	session, err := c.ssh.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var buf bytes.Buffer
	session.Stdout = &buf
	session.Stderr = &buf
	err = session.Run(cmd)
	out := buf.String()
	if err != nil {
		return out, err
	}
	return out, nil
}

// Close 关闭连接
func (c *SFTPClient) Close() error {
	if c.client != nil {
		_ = c.client.Close()
	}
	if c.ssh != nil {
		return c.ssh.Close()
	}
	return nil
}
