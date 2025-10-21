package tray

import (
	"fmt"
	"log"
	"time"

	"github.com/carlosj934/laptop-dashboard-alerting/internal/alerts"
  "github.com/carlosj934/laptop-dashboard-alerting/internal/config"
  "github.com/carlosj934/laptop-dashboard-alerting/internal/metrics"
  "github.com/carlosj934/laptop-dashboard-alerting/internal/notifications"
	"github.com/getlantern/systray"
)

// App reps the system tray application
type App struct {
	config 		*config.Config
	collector *metrics.Collector
	engine 		*alerts.Engine
	notifier	*notifications.Notifier
	running		bool
}

// NewApp creates a new tray app
func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App {
		config: cfg,
		collector: metrics.NewCollector(),
		engine: alerts.NewEngine(cfg),
		notifier: notifications.NewNotifier(),
		running: true,
	}, nil
}

// run starts the system tray app
func (app *App) Run() {
	systray.Run(app.onReady, app.onExit)
}

// onReady is called when the system tray is ready
func (app *App) onReady() {
	// set tray icon and tooltip
	systray.SetTitle("Laptop Monitor")
	systray.SetTooltip("Laptop monitoring and alerts")

	// TODO: set icon - for now just use title
	// systray.SetIcon(iconData)

	// add menu items
	mCurrent := systray.AddMenuItem("View Current Metrics", "Show Current system mtrics")
	systray.AddSeparator()
	mPause := systray.AddMenuItem("Pause Monitoring", "Pause monitoring temporarily")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	// start monitoring in background
	go app.monitorLoop()

	//handle menu clicks
	go func() {
		for {
			select {
			case <-mCurrent.ClickedCh:
				app.showCurrentMetrics()
			case <-mPause.ClickedCh:
				app.toggleMonitoring(mPause)
			case <-mQuit.ClickedCh:
				log.Println("Quitting application...")
				systray.Quit()
				return
			}
		}
	}()

}

// onExit is called when the app exits
func (app *App) onExit() {
	app.running = false
	log.Println("Application exited")
}

// monitorLoop continuously monitors system metrics
func (app *App) monitorLoop() {
	ticker := time.NewTicker(time.Duration(app.config.MonitoringInterval) * time.Second)
	defer ticker.Stop()

	for app.running {
		select {
		case <-ticker.C:
			app.checkMetrics()
		}
	}
}

// checkMetrics collects and evaluates metrics
func (app *App) checkMetrics() {
	// collect metrics
	snapshot, err := app.collector.Collect()
	if err != nil {
		log.Printf("Error collecting metrics: %v", err)
		return
	}

	// evaluate for alerts
	alertEvents := app.engine.Evaluate(snapshot)

	// send notifications for any alerts
	for _, alert := range alertEvents {
		log.Printf("Alert: %s", alert.Message)
		if err := app.notifier.Send(alert); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}
}

// showCurrentMetrics displays current system metrics
func (app *App) showCurrentMetrics() {
	snapshot, err := app.collector.Collect()
	if err != nil {
		log.Printf("Error collecting metrics: %v", err)
		return
	}

	message := fmt.Sprintf(
		"CPU: %.1f%%\nMemory: %.1f%% (%s / %s)\nDisk C: %s free / %s total",
		snapshot.CPUPercent,
    snapshot.MemoryPercent,
    metrics.FormatBytes(snapshot.MemoryUsed),
    metrics.FormatBytes(snapshot.MemoryTotal),
    metrics.FormatBytes(snapshot.DiskFree),
    metrics.FormatBytes(snapshot.DiskTotal),
	)

	// send as notification
	notification := alerts.AlertEvent {
		Timestamp: time.Now(),
		AlertType: "info",
		Message: message,
		CurrentValue: 0,
		ThresholdValue: 0,
	}

	app.notifier.Send(notification)
}

// toggleMonitoring pauses/resumes monitoring
func (app *App) toggleMonitoring(menuItem *systray.MenuItem) {
	if app.running {
		app.running = false
		menuItem.SetTitle("Resume Monitoring")
		log.Println("Monitoring paused")
	} else {
			app.running = true
			menuItem.SetTitle("Pause Monitoring")
			log.Println("Monitoring resumed")
			go app.monitorLoop()
	}
}


