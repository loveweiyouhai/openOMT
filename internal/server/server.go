package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"openOMT/internal/conn"
	"path/filepath"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

const defaultAddr = ":8765"

// Run 启动 HTTP 服务并阻塞
func Run(addr string) error {
	if addr == "" {
		addr = defaultAddr
	}
	mgr := conn.GetManager()

	// 连接
	http.HandleFunc("/api/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var req conn.ConnectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "请求格式错误"})
			return
		}
		if req.Host == "" {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "请填写主机地址"})
			return
		}
		if err := mgr.Connect(req); err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true})
	})

	// 断开
	http.HandleFunc("/api/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = mgr.Close()
		writeJSON(w, map[string]interface{}{"ok": true})
	})

	// 状态
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"connected": mgr.IsConnected(),
			"protocol":  string(mgr.Protocol()),
		})
	})

	// 列表
	http.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		var req conn.ListRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		list, err := mgr.List(req.RemotePath)
		if err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true, "list": list})
	})

	// 下载：GET /api/download?path=xxx
	http.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		remotePath := r.URL.Query().Get("path")
		if remotePath == "" {
			http.Error(w, "缺少 path 参数", http.StatusBadRequest)
			return
		}
		name := filepath.Base(remotePath)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
		if err := mgr.Download(remotePath, w); err != nil {
			log.Printf("download error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 上传：POST /api/upload  multipart: file, remotePath
	http.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		remotePath := r.FormValue("remotePath")
		if remotePath == "" {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "缺少 remotePath"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": "请选择文件: " + err.Error()})
			return
		}
		defer file.Close()
		// 若 remotePath 以 / 结尾或是目录名，则拼上本地文件名
		if strings.HasSuffix(remotePath, "/") || !strings.Contains(filepath.Base(remotePath), ".") {
			remotePath = strings.TrimSuffix(remotePath, "/") + "/" + header.Filename
		}
		if err := mgr.Upload(remotePath, file, header.Size); err != nil {
			writeJSON(w, map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true})
	})

	// 静态资源（嵌入），根路径映射到 static 目录
	root, _ := fs.Sub(staticFS, "static")
	http.Handle("/", http.FileServer(http.FS(root)))

	log.Printf("openOMT 服务已启动: http://127.0.0.1%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
