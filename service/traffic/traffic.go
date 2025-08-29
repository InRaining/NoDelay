package traffic

import (
    "fmt"
    "log"
    "net"
    "sync"
    "time"

    "github.com/InRaining/NoDelay/config"
)

// TrafficLimiterInterface 流量限制器接口
type TrafficLimiterInterface interface {
    CanUseTraffic(playerName string, bytes int64, defaultLimitMB int64) bool
    RecordTraffic(playerName string, bytes int64)
    GetUserInfo(playerName string) (used, limit float64, percentage float64)
    Close()
}

var globalTrafficLimiter TrafficLimiterInterface

// SetGlobalTrafficLimiter 设置全局流量限制器
func SetGlobalTrafficLimiter(limiter TrafficLimiterInterface) {
    globalTrafficLimiter = limiter
}

// CheckUserTrafficByPlayer 检查玩家是否可以使用指定的流量
func CheckUserTrafficByPlayer(playerName string, bytes int64, defaultLimitMB int64) bool {
    if globalTrafficLimiter == nil {
        return true
    }
    return globalTrafficLimiter.CanUseTraffic(playerName, bytes, defaultLimitMB)
}

// RecordUserTrafficByPlayer 记录玩家使用的流量
func RecordUserTrafficByPlayer(playerName string, bytes int64) {
    if globalTrafficLimiter != nil {
        globalTrafficLimiter.RecordTraffic(playerName, bytes)
    }
}

// GetUserTrafficInfoByPlayer 获取玩家的流量使用信息
func GetUserTrafficInfoByPlayer(playerName string) (used, limit float64, percentage float64) {
    if globalTrafficLimiter == nil {
        return 0, 0, 0
    }
    return globalTrafficLimiter.GetUserInfo(playerName)
}

// CheckTrafficLimitByPlayer 在玩家登录时检查流量限制
func CheckTrafficLimitByPlayer(s *config.ConfigProxyService, playerName string) bool {
    if globalTrafficLimiter == nil || !s.Minecraft.EnableTrafficLimit {
        return true
    }

    defaultLimitMB := int64(1024) // 默认1GB
    if s.Minecraft.TrafficLimitMB > 0 {
        defaultLimitMB = s.Minecraft.TrafficLimitMB
    }

    used, limit, percentage := globalTrafficLimiter.GetUserInfo(playerName)

    // 新玩家，允许连接并创建记录
    if used == 0 && limit == 0 {
        return globalTrafficLimiter.CanUseTraffic(playerName, 0, defaultLimitMB)
    }

    // 检查是否超过98%的使用率
    if percentage >= 98.0 {
        log.Printf("Player %s traffic limit exceeded: %.2f MB / %.0f MB (%.1f%%)",
            playerName, used, limit, percentage)
        return false
    }

    return true
}

// AccurateTrafficMonitorConn 是一个net.Conn的包装器，用于精确监控流量
type AccurateTrafficMonitorConn struct {
    net.Conn
    playerName      string
    totalReadBytes  int64
    totalWriteBytes int64
    sessionStart    time.Time
    mutex           sync.Mutex
    service         *config.ConfigProxyService
}

// NewAccurateTrafficMonitorConn 创建一个新的流量监控连接
func NewAccurateTrafficMonitorConn(conn net.Conn, playerName string, s *config.ConfigProxyService) net.Conn {
    return &AccurateTrafficMonitorConn{
        Conn:         conn,
        playerName:   playerName,
        sessionStart: time.Now(),
        service:      s,
    }
}

func (tmc *AccurateTrafficMonitorConn) Read(b []byte) (n int, err error) {
    n, err = tmc.Conn.Read(b)
    if n > 0 {
        tmc.mutex.Lock()
        tmc.totalReadBytes += int64(n)
        total := tmc.totalReadBytes + tmc.totalWriteBytes
        tmc.mutex.Unlock()

        RecordUserTrafficByPlayer(tmc.playerName, int64(n))

        if !CheckUserTrafficByPlayer(tmc.playerName, 0, tmc.service.Minecraft.TrafficLimitMB) {
            tmc.Close()
            return 0, fmt.Errorf("traffic limit exceeded for player %s", tmc.playerName)
        }

        if total%(1024*1024) < int64(n) {
            log.Printf("Traffic Update: %s - Read: %d bytes, Write: %d bytes, Total: %d bytes",
                tmc.playerName, tmc.totalReadBytes, tmc.totalWriteBytes, total)
        }
    }
    return
}

func (tmc *AccurateTrafficMonitorConn) Write(b []byte) (n int, err error) {
    n, err = tmc.Conn.Write(b)
    if n > 0 {
        tmc.mutex.Lock()
        tmc.totalWriteBytes += int64(n)
        total := tmc.totalReadBytes + tmc.totalWriteBytes
        tmc.mutex.Unlock()

        RecordUserTrafficByPlayer(tmc.playerName, int64(n))

        if !CheckUserTrafficByPlayer(tmc.playerName, 0, tmc.service.Minecraft.TrafficLimitMB) {
            tmc.Close()
            return 0, fmt.Errorf("traffic limit exceeded for player %s", tmc.playerName)
        }

        if total%(1024*1024) < int64(n) {
            log.Printf("Traffic Update: %s - Read: %d bytes, Write: %d bytes, Total: %d bytes",
                tmc.playerName, tmc.totalReadBytes, tmc.totalWriteBytes, total)
        }
    }
    return
}

func (tmc *AccurateTrafficMonitorConn) Close() error {
    tmc.mutex.Lock()
    total := tmc.totalReadBytes + tmc.totalWriteBytes
    duration := time.Since(tmc.sessionStart)
    tmc.mutex.Unlock()

    log.Printf("Session ended for %s: Read=%d bytes, Write=%d bytes, Total=%d bytes, Duration=%s",
        tmc.playerName, tmc.totalReadBytes, tmc.totalWriteBytes, total, duration)
    return tmc.Conn.Close()
}