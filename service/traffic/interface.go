package traffic

// 使用 trafficlimiter 包中的类型
type UserTrafficData = trafficlimiter.UserTrafficData

// TrafficLimiterInterface 流量限制器接口
type TrafficLimiterInterface interface {
	CanUseTraffic(playerName string, bytes int64, defaultLimitMB int64) bool
	RecordTraffic(playerName string, bytes int64)
	GetUserInfo(playerName string) (used, limit float64, percentage float64)
	GetAllUsersStats() map[string]UserTrafficData
	ResetUserTraffic(playerName string) bool
	SetUserLimit(playerName string, limitMB int64) bool
	CleanupOldData(cutoffTime int64) bool
	Close()
}

var globalTrafficLimiter TrafficLimiterInterface

// SetTrafficLimiter 设置全局流量限制器
func SetTrafficLimiter(limiter TrafficLimiterInterface) {
	globalTrafficLimiter = limiter
}

// GetTrafficLimiter 获取全局流量限制器
func GetTrafficLimiter() TrafficLimiterInterface {
	return globalTrafficLimiter
}

// CheckUserTraffic 检查用户流量是否可用（基于玩家名称）
func CheckUserTraffic(playerName string, bytes int64, defaultLimitMB int64) bool {
	if globalTrafficLimiter == nil {
		return true
	}
	return globalTrafficLimiter.CanUseTraffic(playerName, bytes, defaultLimitMB)
}

// RecordUserTraffic 记录用户使用的流量（基于玩家名称）
func RecordUserTraffic(playerName string, bytes int64) {
	if globalTrafficLimiter != nil {
		globalTrafficLimiter.RecordTraffic(playerName, bytes)
	}
}

// GetUserTrafficInfo 获取用户流量信息（基于玩家名称）
func GetUserTrafficInfo(playerName string) (used, limit float64, percentage float64) {
	if globalTrafficLimiter == nil {
		return 0, 0, 0
	}
	return globalTrafficLimiter.GetUserInfo(playerName)
}
