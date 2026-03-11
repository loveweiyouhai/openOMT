package conn

// Protocol 连接协议
type Protocol string

const (
	ProtocolSFTP Protocol = "sftp"
	ProtocolFTP  Protocol = "ftp"
)

// ConnectRequest 连接请求
type ConnectRequest struct {
	Protocol Protocol `json:"protocol"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}

// FileInfo 文件/目录信息
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
	IsDir   bool   `json:"isDir"`
}

// ListRequest 列出目录请求
type ListRequest struct {
	RemotePath string `json:"remotePath"`
}
