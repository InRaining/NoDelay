package traffic

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

type UserTrafficData struct {
	PlayerName    string `json:"player_name"`
	UsedBytes     int64  `json:"used_bytes"`
	LimitMB       int64  `json:"limit_mb"`
	LastResetTime int64  `json:"last_reset"`
	LastSeen      int64  `json:"last_seen"`
}

type TrafficLimiter struct {
	dataFile   string
	users      map[string]*UserTrafficData // key: playerName
	mutex      sync.RWMutex
	saveTicker *time.Ticker
	stopChan   chan struct{}
}

func NewTrafficLimiter(dataFile string) *TrafficLimiter {
	tl := &TrafficLimiter{
		dataFile:   dataFile,
		users:      make(map[string]*UserTrafficData),
		saveTicker: time.NewTicker(5 * time.Minute),
		stopChan:   make(chan struct{}),
	}

	tl.loadData()
	go tl.autoSave()
	go tl.autoReset()

	return tl
}

func (tl *TrafficLimiter) loadData() {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	data, err := os.ReadFile(tl.dataFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading traffic data file: %v", err)
		}
		return
	}

	var users map[string]*UserTrafficData
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Printf("Error parsing traffic data: %v", err)
		return
	}

	// Clean up expired data (data older than 7 days)
	cutoff := time.Now().AddDate(0, 0, -7).Unix()
	for _, userData := range users {
		if userData.LastSeen > cutoff {
			tl.users[userData.PlayerName] = userData
		}
	}

	log.Printf("Loaded traffic data for %d players", len(tl.users))
}

func (tl *TrafficLimiter) saveData() {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	data, err := json.MarshalIndent(tl.users, "", "  ")
	if err != nil {
		log.Printf("Error marshaling traffic data: %v", err)
		return
	}

	err = os.WriteFile(tl.dataFile, data, 0644)
	if err != nil {
		log.Printf("Error saving traffic data: %v", err)
	}
}

func (tl *TrafficLimiter) autoSave() {
	for {
		select {
		case <-tl.saveTicker.C:
			tl.saveData()
		case <-tl.stopChan:
			return
		}
	}
}

func (tl *TrafficLimiter) autoReset() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tl.checkAndResetDaily()
		case <-tl.stopChan:
			return
		}
	}
}

func (tl *TrafficLimiter) checkAndResetDaily() {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayUnix := today.Unix()

	resetCount := 0
	for _, userData := range tl.users {
		if userData.LastResetTime < todayUnix {
			userData.UsedBytes = 0
			userData.LastResetTime = todayUnix
			resetCount++
		}
	}

	if resetCount > 0 {
		color.HiCyan("Daily traffic reset: Traffic for %d players has been reset.", resetCount)
	}
}

// CanUseTraffic checks if a player can use the specified amount of traffic.
func (tl *TrafficLimiter) CanUseTraffic(playerName string, bytes int64, defaultLimitMB int64) bool {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	userData, exists := tl.users[playerName]
	if !exists {
		// When creating a new user, set the correct limit immediately.
		userData = &UserTrafficData{
			PlayerName:    playerName,
			UsedBytes:     0,
			LimitMB:       defaultLimitMB, // Ensure the passed default limit is used.
			LastResetTime: time.Now().Unix(),
			LastSeen:      time.Now().Unix(),
		}
		tl.users[playerName] = userData
		log.Printf("Created new player traffic record: %s (Limit: %d MB)", playerName, defaultLimitMB)
	}

	// Update the last seen time.
	userData.LastSeen = time.Now().Unix()

	// Check if a reset is needed (new day).
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if userData.LastResetTime < today.Unix() {
		userData.UsedBytes = 0
		userData.LastResetTime = today.Unix()
		log.Printf("Reset daily traffic for player: %s", playerName)
	}

	// If the user's limit is 0, update it to the new default limit.
	if userData.LimitMB == 0 && defaultLimitMB > 0 {
		userData.LimitMB = defaultLimitMB
		log.Printf("Updated player %s limit from 0 to %d MB", playerName, defaultLimitMB)
	}

	limitBytes := userData.LimitMB * 1024 * 1024
	canUse := userData.UsedBytes+bytes <= limitBytes

	return canUse
}

// RecordTraffic records the traffic used by a player.
func (tl *TrafficLimiter) RecordTraffic(playerName string, bytes int64) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	userData, exists := tl.users[playerName]
	if !exists {
		return
	}

	userData.UsedBytes += bytes
	userData.LastSeen = time.Now().Unix()
}

// GetUserInfo gets player traffic information.
func (tl *TrafficLimiter) GetUserInfo(playerName string) (used, limit float64, percentage float64) {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	userData, exists := tl.users[playerName]
	if !exists {
		// If the user does not exist, return 0 without creating a record.
		return 0, 0, 0
	}

	used = float64(userData.UsedBytes) / (1024 * 1024) // Convert to MB
	limit = float64(userData.LimitMB)
	if limit > 0 {
		percentage = (used / limit) * 100
	}

	return used, limit, percentage
}

// GetAllUsersStats returns all user traffic statistics.
func (tl *TrafficLimiter) GetAllUsersStats() map[string]UserTrafficData {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	stats := make(map[string]UserTrafficData)
	for playerName, userData := range tl.users {
		stats[playerName] = *userData
	}

	return stats
}

// ResetUserTraffic resets the traffic for a specific player.
func (tl *TrafficLimiter) ResetUserTraffic(playerName string) bool {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	userData, exists := tl.users[playerName]
	if !exists {
		return false
	}

	userData.UsedBytes = 0
	userData.LastResetTime = time.Now().Unix()
	return true
}

// SetUserLimit sets the traffic limit for a player.
func (tl *TrafficLimiter) SetUserLimit(playerName string, limitMB int64) bool {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	userData, exists := tl.users[playerName]
	if !exists {
		userData = &UserTrafficData{
			PlayerName:    playerName,
			UsedBytes:     0,
			LimitMB:       limitMB,
			LastResetTime: time.Now().Unix(),
			LastSeen:      time.Now().Unix(),
		}
		tl.users[playerName] = userData
	} else {
		userData.LimitMB = limitMB
		userData.LastSeen = time.Now().Unix()
	}

	return true
}

// CleanupOldData cleans up expired data.
func (tl *TrafficLimiter) CleanupOldData(cutoffTime int64) bool {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	initialCount := len(tl.users)
	for playerName, userData := range tl.users {
		if userData.LastSeen < cutoffTime {
			delete(tl.users, playerName)
		}
	}

	finalCount := len(tl.users)
	removed := initialCount - finalCount

	if removed > 0 {
		color.HiGreen("Cleaned up data for %d expired players.", removed)
		return true
	}

	return false
}

// Close shuts down the traffic limiter.
func (tl *TrafficLimiter) Close() {
	close(tl.stopChan)
	tl.saveTicker.Stop()
	tl.saveData()
	color.HiGreen("Traffic data saved.")
}