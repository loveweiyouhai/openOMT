package conn

import "errors"

var (
	ErrNotConnected           = errors.New("未连接服务器")
	ErrUnsupportedProtocol    = errors.New("不支持的协议")
	ErrTerminalNotSupported   = errors.New("仅 SFTP (SSH) 连接支持远程终端，当前为 FTP 连接")
)
