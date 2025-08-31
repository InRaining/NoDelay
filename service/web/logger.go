package web

import (
    "bytes"
    "container/ring"
    "embed"
    "fmt"
    "io"
    "log"
    "net/http"
    "sync"

    "github.com/InRaining/NoDelay/config"
    "github.com/fatih/color"
)

//go:embed index.html
var webContent embed.FS

const (
    logBufferSize = 512 // 存储最新的 2048 条日志
    logChannelCap = 64  // 日志消息通道的缓冲容量
)

// Logger 捕获日志输出以便在web上显示
type Logger struct {
    logChan      chan []byte
    originalOut  io.Writer
    logRing      *ring.Ring
    mu           sync.RWMutex
    processDone  chan struct{}
}

var webLogger *Logger

// NewLogger 创建一个新的 Logger 实例并启动后台处理 goroutine
func NewLogger(originalWriter io.Writer) *Logger {
    webLogger = &Logger{
        logChan:     make(chan []byte, logChannelCap),
        originalOut: originalWriter,
        logRing:     ring.New(logBufferSize),
        processDone: make(chan struct{}),
    }
    go webLogger.processLogs()
    return webLogger
}

// Write 实现了 io.Writer 接口。它将日志消息发送到 channel，是非阻塞的。
func (l *Logger) Write(p []byte) (n int, err error) {
    // 必须复制 p，因为底层的字节数组可能会被重用
    msg := make([]byte, len(p))
    copy(msg, p)

    select {
    case l.logChan <- msg:
        // 成功发送
    default:
        // channel 已满，丢弃日志以防止阻塞
    }
    return len(p), nil
}

// processLogs 是一个后台 goroutine，负责处理 channel 中的日志消息
func (l *Logger) processLogs() {
    for msg := range l.logChan {
        // 1. 写入原始输出 (e.g., console)
        l.originalOut.Write(msg)

        // 2. 写入环形缓冲区以供 web 显示
        l.mu.Lock()
        l.logRing.Value = msg
        l.logRing = l.logRing.Next()
        l.mu.Unlock()
    }
    close(l.processDone)
}

// Close 安全地关闭 logger
func (l *Logger) Close() {
    close(l.logChan)
    <-l.processDone // 等待 processLogs goroutine 结束
}

// getLogsAsString 从环形缓冲区中获取所有日志并格式化为字符串
func (l *Logger) getLogsAsString() string {
    l.mu.RLock()
    defer l.mu.RUnlock()

    var b bytes.Buffer
    l.logRing.Do(func(p interface{}) {
        if p != nil {
            b.Write(p.([]byte))
        }
    })
    return b.String()
}

// StartWebServer 启动用于显示日志的HTTP服务器
func StartWebServer() {
    port := "8088"
    if config.Config.WebLogPort > 0 {
        port = fmt.Sprintf("%d", config.Config.WebLogPort)
    }

    addr := "0.0.0.0:" + port
    mux := http.NewServeMux()

    // 根路径 "/" 提供 index.html
    mux.Handle("/", http.FileServer(http.FS(webContent)))
    // "/logs" 路径提供纯文本日志数据
    mux.HandleFunc("/logs", logsApiHandler)

    log.Printf(color.HiCyanString("Starting web log server on http://%s", addr))

    go func() {
        if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
            log.Printf(color.HiRedString("Web log server error: %v", err))
        }
    }()
}

// logsApiHandler 提供原始日志数据
func logsApiHandler(w http.ResponseWriter, r *http.Request) {
    if webLogger == nil {
        http.Error(w, "Logger not initialized", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    fmt.Fprint(w, webLogger.getLogsAsString())
}