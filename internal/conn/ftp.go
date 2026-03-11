package conn

import (
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPClient FTP 连接封装
type FTPClient struct {
	conn *ftp.ServerConn
}

// NewFTPClient 创建 FTP 客户端
func NewFTPClient(req ConnectRequest) (*FTPClient, error) {
	port := req.Port
	if port <= 0 {
		port = 21
	}
	addr := fmt.Sprintf("%s:%d", req.Host, port)
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(15*time.Second))
	if err != nil {
		return nil, fmt.Errorf("ftp dial: %w", err)
	}
	err = conn.Login(req.Username, req.Password)
	if err != nil {
		conn.Quit()
		return nil, fmt.Errorf("ftp login: %w", err)
	}
	return &FTPClient{conn: conn}, nil
}

// List 列出目录
func (c *FTPClient) List(remotePath string) ([]FileInfo, error) {
	if remotePath != "" && remotePath != "." {
		if err := c.conn.ChangeDir(remotePath); err != nil {
			return nil, err
		}
	}
	entries, err := c.conn.List(".")
	if err != nil {
		return nil, err
	}
	var list []FileInfo
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		fullPath := path.Join(remotePath, e.Name)
		if remotePath == "" {
			fullPath = e.Name
		}
		list = append(list, FileInfo{
			Name:    e.Name,
			Path:    fullPath,
			Size:    int64(e.Size),
			ModTime: e.Time.Format("2006-01-02 15:04:05"),
			IsDir:   e.Type == ftp.EntryTypeFolder,
		})
	}
	return list, nil
}

// Download 下载文件到 Writer
func (c *FTPClient) Download(remotePath string, w io.Writer) error {
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return err
	}
	defer resp.Close()
	_, err = io.Copy(w, resp)
	return err
}

// Upload 从 Reader 上传文件
func (c *FTPClient) Upload(remotePath string, r io.Reader, size int64) error {
	return c.conn.Stor(remotePath, r)
}

// MkdirAll 创建远程目录（FTP 仅支持单级，多级需逐级创建）
func (c *FTPClient) MkdirAll(remotePath string) error {
	remotePath = path.Clean(remotePath)
	if remotePath == "" || remotePath == "." {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(remotePath, "/"), "/")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		p := "/" + strings.Join(parts[:i+1], "/")
		if err := c.conn.MakeDir(p); err != nil {
			// 已存在则忽略
			if !strings.Contains(strings.ToLower(err.Error()), "exists") && !strings.Contains(err.Error(), "550") {
				return err
			}
		}
	}
	return nil
}

// CreateEmptyFile 创建空文件
func (c *FTPClient) CreateEmptyFile(remotePath string) error {
	return c.conn.Stor(remotePath, strings.NewReader(""))
}

// DeleteFile 删除单个文件
func (c *FTPClient) DeleteFile(remotePath string) error {
	return c.conn.Delete(remotePath)
}

// DeleteDirRecursive 递归删除目录及其内容
func (c *FTPClient) DeleteDirRecursive(remotePath string) error {
	return c.conn.RemoveDirRecur(remotePath)
}

// Close 关闭连接
func (c *FTPClient) Close() error {
	if c.conn != nil {
		return c.conn.Quit()
	}
	return nil
}

// parseFTPList 解析 LIST 行（备用）
func parseFTPList(line string) (name string, size int64, isDir bool) {
	parts := strings.Fields(line)
	if len(parts) < 9 {
		return "", 0, false
	}
	isDir = strings.HasPrefix(parts[0], "d")
	size, _ = strconv.ParseInt(parts[4], 10, 64)
	name = strings.Join(parts[8:], " ")
	return name, size, isDir
}
