package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"
	"io/ioutil"

	"github.com/InRaining/NoDelay/config"
	"github.com/InRaining/NoDelay/console"
	"github.com/InRaining/NoDelay/service"
	"github.com/InRaining/NoDelay/version"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

func main() {
	log.SetOutput(color.Output)
	console.SetTitle(fmt.Sprintf("NoDelay %v | Running...", version.Version))
	color.HiGreen("Welcome to NoDelay %s (%s)!\n", version.Version, version.CommitHash)
	color.HiGreen("Developer: InRaining")
	color.HiGreen("Repository: https://github.com/InRaining/NoDelay")

	color.HiBlue("Please wait for 5 seconds for verification...")
	time.AfterFunc(5*time.Second, func() {
		resp, err := http.Get("https://bind.hln.asia/NoDelay/NoDelay.php")
		if err != nil {
			log.Panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}

		if string(body) == "true" {
			color.HiGreen("You are currently allowed to use NoDelay: your IP is authorized.")
				config.LoadConfig()
				service.Listeners = make([]net.Listener, 0, len(config.Config.Services))

				{
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						log.Panic(err)
					}
					defer watcher.Close()
					err = monitorConfig(watcher)
					if err != nil {
						log.Panic("Config Reload Error : ", err)
					}
				}

				{
					osSignals := make(chan os.Signal, 1)
					signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
					<-osSignals
					// stop the program
					service.CleanupServices()
				}
			} else {
				color.HiRed("You are currently not allowed to run NoDelay: your IP is not authorized.")
			}
		})

	select {}
}

func monitorConfig(watcher *fsnotify.Watcher) error {
	ctx, cancel := context.WithCancel(context.Background())
	service.ExecuteServices(ctx)
	go func() {
		reloadSignal := make(chan os.Signal, 1)
		signal.Notify(reloadSignal, syscall.SIGHUP)
		defer signal.Stop(reloadSignal)
		for {
			select {
			case _, ok := <-reloadSignal:
				if !ok {
					log.Println(color.HiRedString("Config Reload Error : Signal channel unexpectedly closed"))
					return
				}

			case event, ok := <-watcher.Events:
				if !ok {
					log.Println(color.HiRedString("Config Reload Error : Watcher event channel unexpectedly closed"))
					return
				}
				if event.Op.Has(fsnotify.Write) { // config reload
					// wait for the file to finish writing
					timer := time.NewTimer(100 * time.Millisecond)
					for {
						select {
						case <-watcher.Events:
							timer.Reset(100 * time.Millisecond)
						case <-timer.C:
							goto reload
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println(color.HiRedString("Config Reload Error : Watcher error channel unexpectedly closed"))
					return
				}
				log.Println(color.HiRedString("Config Reload Error : ", err))
				return
			}
		reload:
			log.Println(color.HiMagentaString("Config Reload : File change detected. Reloading..."))
			if config.LoadLists(true) { // reload success
				log.Println(color.HiMagentaString("Config Reload : Successfully reloaded Lists."))
				cancel()
				service.CleanupServices()
				service.Listeners = make([]net.Listener, 0, len(config.Config.Services))
				ctx, cancel = context.WithCancel(context.Background())
				service.ExecuteServices(ctx)
			} else {
				log.Println(color.HiMagentaString("Config Reload : Failed to reload Lists."))
			}
		}
	}()
	return watcher.Add("NoDelay.json")
}
