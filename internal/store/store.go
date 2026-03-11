package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"openFTP/internal/conn"
)

const keyFile = "key.bin"
const dbFile = "connections.db"
const aesKeySize = 32

// SavedConnection 已保存的连接（列表项，不含密码）
type SavedConnection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Protocol string `json:"protocol"`
}

type Store struct {
	mu  sync.Mutex
	dir string
	key []byte
	db  *sql.DB
}

// New 创建存储，dir 为空则使用用户配置目录
func New(configDir string) (*Store, error) {
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			dir = "."
		}
		configDir = filepath.Join(dir, "OpenFTP")
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}
	s := &Store{dir: configDir}
	if err := s.initKey(); err != nil {
		return nil, err
	}
	if err := s.initDB(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) initKey() error {
	keyPath := filepath.Join(s.dir, keyFile)
	data, err := os.ReadFile(keyPath)
	if err == nil && len(data) == aesKeySize {
		s.key = data
		return nil
	}
	s.key = make([]byte, aesKeySize)
	if _, err := io.ReadFull(rand.Reader, s.key); err != nil {
		return err
	}
	return os.WriteFile(keyPath, s.key, 0600)
}

func (s *Store) initDB() error {
	dbPath := filepath.Join(s.dir, dbFile)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	s.db = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id TEXT PRIMARY KEY,
			name TEXT,
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			username TEXT,
			password TEXT,
			protocol TEXT DEFAULT 'sftp',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (s *Store) encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Store) decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// List 返回已保存连接列表（不含密码）
func (s *Store) List() ([]SavedConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT id, name, host, port, username, protocol FROM connections ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SavedConnection
	for rows.Next() {
		var c SavedConnection
		var name sql.NullString
		if err := rows.Scan(&c.ID, &name, &c.Host, &c.Port, &c.Username, &c.Protocol); err != nil {
			return nil, err
		}
		if name.Valid && name.String != "" {
			c.Name = name.String
		} else {
			c.Name = c.Host
		}
		if c.Protocol == "" {
			if c.Port == 21 {
				c.Protocol = "ftp"
			} else {
				c.Protocol = "sftp"
			}
		}
		list = append(list, c)
	}
	return list, nil
}

// GetByID 按 ID 获取完整连接（含解密密码），用于连接
func (s *Store) GetByID(id string) (conn.ConnectRequest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var host, username, encPassword, protocol string
	var port int
	err := s.db.QueryRow(`SELECT host, port, username, password, protocol FROM connections WHERE id = ?`, id).
		Scan(&host, &port, &username, &encPassword, &protocol)
	if err != nil {
		return conn.ConnectRequest{}, false
	}

	password, _ := s.decrypt(encPassword)

	if protocol == "" {
		if port == 21 {
			protocol = "ftp"
		} else {
			protocol = "sftp"
		}
	}

	return conn.ConnectRequest{
		Protocol: conn.Protocol(protocol),
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}, true
}

// Save 新增或更新一条连接
func (s *Store) Save(id, name, host string, port int, username, password, protocol string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if port <= 0 {
		port = 22
	}
	if protocol == "" {
		protocol = "sftp"
	}

	encPassword, err := s.encrypt(password)
	if err != nil {
		return "", err
	}

	if id != "" {
		var exists int
		s.db.QueryRow(`SELECT COUNT(*) FROM connections WHERE id = ?`, id).Scan(&exists)
		if exists > 0 {
			_, err = s.db.Exec(`
				UPDATE connections SET name=?, host=?, port=?, username=?, password=?, protocol=?, updated_at=CURRENT_TIMESTAMP 
				WHERE id=?`,
				name, host, port, username, encPassword, protocol, id)
			return id, err
		}
	}

	newID := fmt.Sprintf("conn-%d", s.getNextID())
	_, err = s.db.Exec(`
		INSERT INTO connections (id, name, host, port, username, password, protocol) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newID, name, host, port, username, encPassword, protocol)
	return newID, err
}

func (s *Store) getNextID() int {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM connections`).Scan(&count)
	return count + 1
}

// Delete 删除一条连接
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM connections WHERE id = ?`, id)
	return err
}

// Rename 仅修改显示名
func (s *Store) Rename(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`UPDATE connections SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, name, id)
	return err
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
