package conn

import (
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/ssh"
)

// ShellSession 交互式 shell 会话
type ShellSession struct {
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
	mu      sync.Mutex
	closed  bool
}

// NewShellSession 创建新的交互式 shell 会话
func NewShellSession(sshClient *ssh.Client) (*ShellSession, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	// 请求伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("请求 PTY 失败: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("获取 stdin 失败: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("获取 stdout 失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("获取 stderr 失败: %w", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("启动 shell 失败: %w", err)
	}

	return &ShellSession{
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}, nil
}

// Write 向 shell 写入数据
func (s *ShellSession) Write(data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("会话已关闭")
	}
	_, err := s.stdin.Write([]byte(data))
	return err
}

// Resize 调整终端大小
func (s *ShellSession) Resize(rows, cols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("会话已关闭")
	}
	return s.session.WindowChange(rows, cols)
}

// ReadOutput 读取输出（包含stdout和stderr的合并读取器）
func (s *ShellSession) ReadOutput(buf []byte) (int, error) {
	return s.stdout.Read(buf)
}

// ReadStderr 读取 stderr
func (s *ShellSession) ReadStderr(buf []byte) (int, error) {
	return s.stderr.Read(buf)
}

// Close 关闭会话
func (s *ShellSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}
