package traffic

import (
    "fmt"
    "log"
    "net"
    "sync"
    "syscall"
    "time"

    "github.com/InRaining/NoDelay/config"
)

// CheckUserTrafficByPlayer checks if a player can use the specified amount of traffic.
func CheckUserTrafficByPlayer(playerName string, bytes int64, defaultLimitMB int64) bool {
    if globalTrafficLimiter == nil {
        return true
    }
    return globalTrafficLimiter.CanUseTraffic(playerName, bytes, defaultLimitMB)
}

// RecordUserTrafficByPlayer records the traffic used by a player.
func RecordUserTrafficByPlayer(playerName string, bytes int64) {
    if globalTrafficLimiter != nil {
        globalTrafficLimiter.RecordTraffic(playerName, bytes)
    }
}

// GetUserTrafficInfoByPlayer gets player traffic information.
func GetUserTrafficInfoByPlayer(playerName string) (used, limit float64, percentage float64) {
    if globalTrafficLimiter == nil {
        return 0, 0, 0
    }
    return globalTrafficLimiter.GetUserInfo(playerName)
}

// CheckTrafficLimitByPlayer checks the traffic limit for a player upon login.
func CheckTrafficLimitByPlayer(s *config.ConfigProxyService, playerName string) bool {
    // Traffic limit settings are global, not per-service.
    // Access them from the global config.
    if globalTrafficLimiter == nil || config.Config.TrafficLimiter == nil || !config.Config.TrafficLimiter.EnableTrafficLimit {
        return true
    }

    if config.Config.TrafficLimiter.TrafficLimitMB > 0 {
        defaultLimitMB = config.Config.TrafficLimiter.TrafficLimitMB
    } else {
        defaultLimitMB := int64(1024)
    }

    used, limit, percentage := globalTrafficLimiter.GetUserInfo(playerName)

    // For new players, allow connection and create a record.
    if used == 0 && limit == 0 {
        return globalTrafficLimiter.CanUseTraffic(playerName, 0, defaultLimitMB)
    }
    
    if percentage >= 98.0 {
        log.Printf("Player %s traffic limit exceeded: %.2f MB / %.0f MB (%.1f%%)",
            playerName, used, limit, percentage)
        return false
    }
    
    return true
}

// AccurateTrafficMonitorConn is a wrapper for net.Conn to accurately monitor traffic.
type AccurateTrafficMonitorConn struct {
    net.Conn
    playerName      string
    totalReadBytes  int64
    totalWriteBytes int64
    sessionStart    time.Time
    mutex           sync.Mutex
    service         *config.ConfigProxyService
}

// NewAccurateTrafficMonitorConn creates a new traffic monitoring connection.
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

        // The traffic limit is global, so get it from the global config.
        if config.Config.TrafficLimiter != nil && !CheckUserTrafficByPlayer(tmc.playerName, 0, config.Config.TrafficLimiter.TrafficLimitMB) {
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

        // The traffic limit is global, so get it from the global config.
        if config.Config.TrafficLimiter != nil && !CheckUserTrafficByPlayer(tmc.playerName, 0, config.Config.TrafficLimiter.TrafficLimitMB) {
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

func (tmc *AccurateTrafficMonitorConn) SyscallConn() (syscall.RawConn, error) {
    sc, ok := tmc.Conn.(syscall.Conn)
    if !ok {
        return nil, fmt.Errorf("underlying connection does not implement syscall.Conn")
    }
    return sc.SyscallConn()
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
