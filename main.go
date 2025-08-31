package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/InRaining/NoDelay/config"
	"github.com/InRaining/NoDelay/console"
	"github.com/InRaining/NoDelay/service"
	"github.com/InRaining/NoDelay/service/traffic"
	"github.com/InRaining/NoDelay/version"
	"github.com/InRaining/NoDelay/service/web"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

var (
	trafficLimiter traffic.TrafficLimiterInterface
	webLogger      *web.Logger
)

const authURL = "https://bind.hln.asia/NoDelay/NoDelay.php"

func main() {
	webLogger = web.NewLogger(color.Output)
    log.SetOutput(webLogger)
    color.SetOutput(webLogger)

	console.SetTitle(fmt.Sprintf("NoDelay %v | Running...", version.Version))
	color.HiGreen("Welcome to NoDelay %s (%s)!", version.Version, version.CommitHash)
	color.HiGreen("Developer: InRaining")
	color.HiGreen("Repository: https://github.com/InRaining/NoDelay")

	color.HiBlue("Verifying authorization...")
	if !checkAuth() {
		color.HiRed("Authorization failed. Your IP is not authorized. The program will exit in 3 seconds.")
		time.Sleep(3 * time.Second)
		return
	}
	color.HiGreen("Authorization successful. Starting services...")

	if err := startup(); err != nil {
		log.Panic(err)
	}

	waitForShutdown()
	cleanup()
}

func checkAuth() bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(authURL)
	if err != nil {
		log.Printf(color.HiRedString("Authorization request failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf(color.HiRedString("Authorization server returned error status: %s", resp.Status))
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(color.HiRedString("Failed to read authorization response: %v", err))
		return false
	}

	return string(body) == "true"
}

func startup() error {
	config.LoadConfig()

	web.StartWebServer()

    if _, err := os.Stat("TrafficTable.json"); os.IsNotExist(err) {
        log.Println(color.HiYellowString("Traffic Table is not exists. Generating a new one..."))
        if err := os.WriteFile("TrafficTable.json", []byte("{}"), 0644); err != nil {
            return fmt.Errorf("failed to create Traffic Table: %w", err)
        }
    }
	initTrafficLimiter()

	service.Listeners = make([]net.Listener, 0, len(config.Config.Services))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	go monitorConfig(watcher)

	return nil
}

func waitForShutdown() {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals
}

func initTrafficLimiter() {
	color.HiCyan("Initializing traffic limiter...")
	limiter := traffic.NewTrafficLimiter("TrafficTable.json")
	trafficLimiter = limiter
	traffic.SetGlobalTrafficLimiter(limiter)
	color.HiGreen("Traffic limiter initialized.")
	go startTrafficStatsDisplay()
}

func startTrafficStatsDisplay() {
	displayTrafficStats()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		displayTrafficStats()
	}
}

func displayTrafficStats() {
	if trafficLimiter == nil {
		return
	}

	stats := trafficLimiter.GetAllUsersStats()
	if len(stats) == 0 {
		return
	}

	players := make([]string, 0, len(stats))
	for player := range stats {
		players = append(players, player)
	}
	sort.Strings(players)

	color.HiCyan("\n---------- Current Traffic Usage Stats (%s) ----------", time.Now().Format("15:04:05"))
	color.HiCyan("%-16s %-12s %-12s %-10s", "Player", "Used", "Limit", "Usage")
	color.White("---------------------------------------------------------")

	for _, player := range players {
		stat := stats[player]
		usedMB := float64(stat.UsedBytes) / (1024 * 1024)
		limitMB := float64(stat.LimitMB)

		var percentage float64
		if limitMB > 0 {
			percentage = (usedMB / limitMB) * 100
		}

		var statusColor func(format string, a ...interface{}) string
		switch {
		case percentage > 90:
			statusColor = color.HiRedString
		case percentage > 70:
			statusColor = color.HiYellowString
		default:
			statusColor = color.HiGreenString
		}

		fmt.Println(statusColor("%-16s %8.2f MB %8.0f MB %9.1f%%",
			player, usedMB, limitMB, percentage))
	}
	color.White("---------------------------------------------------------\n")
}

func monitorConfig(watcher *fsnotify.Watcher) {
	defer watcher.Close()

	ctx, cancel := context.WithCancel(context.Background())
	service.ExecuteServices(ctx)

	if err := watcher.Add("NoDelay.json"); err != nil {
		log.Println(color.HiRedString("Failed to watch config file: %v", err))
	}
    if err := watcher.Add("TrafficTable.json"); err != nil {
        log.Println(color.HiRedString("Failed to watch traffic data file: %v", err))
    }

	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	defer signal.Stop(reloadSignal)

	configReloadTimer := time.NewTimer(time.Hour)
    configReloadTimer.Stop()
    trafficReloadTimer := time.NewTimer(time.Hour)
    trafficReloadTimer.Stop()

    for {
        select {
        case <-reloadSignal:
            goto reloadConfig

        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            if event.Has(fsnotify.Write) {
                switch event.Name {
                case "NoDelay.json":
                    configReloadTimer.Reset(100 * time.Millisecond)
                case "TrafficTable.json":
                    trafficReloadTimer.Reset(100 * time.Millisecond)
                }
            }

        case <-configReloadTimer.C:
            goto reloadConfig

        case <-trafficReloadTimer.C:
            goto reloadTraffic

        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Println(color.HiRedString("Error while watching config: ", err))
            return
        }
        continue

		reloadConfig:
			log.Println(color.HiMagentaString("File change detected, reloading configuration..."))
			if config.LoadLists(true) {
				log.Println(color.HiMagentaString("Lists reloaded successfully."))
				cancel()
				service.CleanupServices()
				service.Listeners = make([]net.Listener, 0, len(config.Config.Services))
				ctx, cancel = context.WithCancel(context.Background())
				service.ExecuteServices(ctx)
			} else {
				log.Println(color.HiRedString("Failed to reload config."))
			}
			continue
		
		reloadTraffic:
        	if trafficLimiter != nil {
            	trafficLimiter.ReloadData()
        	}
        	continue
	}
}

func cleanup() {
	color.HiYellow("Shutting down services...")
	service.CleanupServices()

	if trafficLimiter != nil {
		color.HiYellow("Saving traffic data...")
		trafficLimiter.Close()
		color.HiGreen("Traffic data saved.")
	}

	color.HiGreen("Services have been shut down.")

	if webLogger != nil {
        webLogger.Close()
    }
	
	os.Exit(0)
}