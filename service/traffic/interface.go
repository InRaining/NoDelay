package traffic

// UserTrafficData holds the traffic usage data for a single player.
// It is defined here because it's part of the public interface contract.
type UserTrafficData struct {
    PlayerName    string `json:"player_name"`
    UsedBytes     int64  `json:"used_bytes"`
    LimitMB       int64  `json:"limit_mb"`
    LastResetTime int64  `json:"last_reset"`
    LastSeen      int64  `json:"last_seen"`
}

// TrafficLimiterInterface defines the interface for traffic limiters.
type TrafficLimiterInterface interface {
    CanUseTraffic(playerName string, bytes int64, defaultLimitMB int64) bool
    RecordTraffic(playerName string, bytes int64)
    GetUserInfo(playerName string) (used, limit float64, percentage float64)
    Close()
    GetAllUsersStats() map[string]UserTrafficData
    ResetUserTraffic(playerName string) bool
    SetUserLimit(playerName string, limitMB int64) bool
    CleanupOldData(cutoffTime int64) bool
}

var globalTrafficLimiter TrafficLimiterInterface

// SetGlobalTrafficLimiter sets the global traffic limiter instance.
func SetGlobalTrafficLimiter(limiter TrafficLimiterInterface) {
    globalTrafficLimiter = limiter
}