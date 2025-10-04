package config

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigWatcher manages hot-reload functionality for configuration files
type ConfigWatcher struct {
	viper     *viper.Viper
	watcher   *fsnotify.Watcher
	config    *Config
	callbacks []ReloadCallback
	debouncer *Debouncer
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// ReloadCallback is called when configuration is successfully reloaded
type ReloadCallback func(oldConfig, newConfig *Config) error

// Debouncer prevents rapid-fire reloads when files are modified multiple times quickly
type Debouncer struct {
	timer    *time.Timer
	duration time.Duration
	mu       sync.Mutex
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{
		duration: duration,
	}
}

// Debounce resets the timer and executes the function after the delay
func (d *Debouncer) Debounce(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.duration, fn)
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(v *viper.Viper, initialConfig *Config) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cw := &ConfigWatcher{
		viper:     v,
		watcher:   watcher,
		config:    initialConfig,
		callbacks: make([]ReloadCallback, 0),
		debouncer: NewDebouncer(500 * time.Millisecond), // 500ms debounce delay
		ctx:       ctx,
		cancel:    cancel,
	}

	return cw, nil
}

// AddReloadCallback registers a callback to be executed when config reloads
func (cw *ConfigWatcher) AddReloadCallback(callback ReloadCallback) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.callbacks = append(cw.callbacks, callback)
}

// Start begins watching configuration files for changes
func (cw *ConfigWatcher) Start() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.isRunning {
		return fmt.Errorf("config watcher is already running")
	}

	// Add configuration paths to watcher
	configPaths := []string{"./configs", "."}
	for _, path := range configPaths {
		if err := cw.watcher.Add(path); err != nil {
			// If path doesn't exist, continue with other paths
			continue
		}
	}

	cw.isRunning = true

	// Start the watcher goroutine
	go cw.watchLoop()

	return nil
}

// Stop stops the configuration watcher
func (cw *ConfigWatcher) Stop() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.isRunning {
		return nil
	}

	cw.cancel()
	cw.isRunning = false

	return cw.watcher.Close()
}

// GetCurrentConfig returns the current configuration (thread-safe)
func (cw *ConfigWatcher) GetCurrentConfig() *Config {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

// watchLoop is the main event loop for file watching
func (cw *ConfigWatcher) watchLoop() {
	for {
		select {
		case <-cw.ctx.Done():
			return

		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Only process write and create events for config files
			if cw.isConfigFile(event.Name) && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				cw.debouncer.Debounce(func() {
					if err := cw.reloadConfig(); err != nil {
						// In production, you might want to log this error
						// For now, we'll continue running even if reload fails
						fmt.Printf("Config reload failed: %v\n", err)
					}
				})
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			// In production, you might want to log this error
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}
}

// isConfigFile checks if the file is a configuration file we care about
func (cw *ConfigWatcher) isConfigFile(filename string) bool {
	configFiles := []string{
		"config.yaml", "config.yml", "config.json",
		"./configs/config.yaml", "./configs/config.yml", "./configs/config.json",
		"configs/config.yaml", "configs/config.yml", "configs/config.json",
	}

	for _, configFile := range configFiles {
		if strings.HasSuffix(filename, configFile) {
			return true
		}
	}

	return false
}

// reloadConfig reloads the configuration from file and validates it
func (cw *ConfigWatcher) reloadConfig() error {
	// Read the updated configuration
	if err := cw.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Create a new config instance
	newConfig := &Config{}
	if err := cw.viper.Unmarshal(newConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the new configuration
	if err := validate.Struct(newConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Execute callbacks with the old and new configuration
	cw.mu.RLock()
	oldConfig := cw.config
	callbacks := make([]ReloadCallback, len(cw.callbacks))
	copy(callbacks, cw.callbacks)
	cw.mu.RUnlock()

	// Execute all callbacks
	for _, callback := range callbacks {
		if err := callback(oldConfig, newConfig); err != nil {
			return fmt.Errorf("reload callback failed: %w", err)
		}
	}

	// Update the current configuration
	cw.mu.Lock()
	cw.config = newConfig
	cw.mu.Unlock()

	return nil
}

// EnableHotReload creates and starts a config watcher if hot-reload is enabled
func EnableHotReload(v *viper.Viper, config *Config) (*ConfigWatcher, error) {
	if !config.EnableHotReload {
		return nil, nil // Hot-reload disabled, return nil watcher
	}

	watcher, err := NewConfigWatcher(v, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create config watcher: %w", err)
	}

	if err := watcher.Start(); err != nil {
		return nil, fmt.Errorf("failed to start config watcher: %w", err)
	}

	return watcher, nil
}

// HotReloadConfig wraps the standard NewConfig with hot-reload capability
func HotReloadConfig() (*Config, *ConfigWatcher, error) {
	// Create initial configuration using the standard NewConfig
	config, err := NewConfig()
	if err != nil {
		return nil, nil, err
	}

	// Create viper instance for hot-reload (similar to NewConfig setup)
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	// Bind environment variables
	v.BindEnv("DB_DSN")
	v.BindEnv("JWT_SECRET")
	v.BindEnv("LOG_LEVEL")
	v.BindEnv("ENABLE_HOT_RELOAD")

	// Set defaults (same as NewConfig)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("ENABLE_HOT_RELOAD", false)

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional, continue with env vars and defaults
	}

	// Enable hot-reload if configured
	watcher, err := EnableHotReload(v, config)
	if err != nil {
		return config, nil, fmt.Errorf("failed to enable hot-reload: %w", err)
	}

	return config, watcher, nil
}

// ConfigWithWatcher represents a configuration with optional hot-reload watcher
type ConfigWithWatcher struct {
	Config  *Config
	Watcher *ConfigWatcher
}

// Close properly shuts down the watcher if it exists
func (cw *ConfigWithWatcher) Close() error {
	if cw.Watcher != nil {
		return cw.Watcher.Stop()
	}
	return nil
}

// GetConfig returns the current configuration, checking the watcher if available
func (cw *ConfigWithWatcher) GetConfig() *Config {
	if cw.Watcher != nil {
		return cw.Watcher.GetCurrentConfig()
	}
	return cw.Config
}

// NewConfigWithHotReload creates a new configuration with optional hot-reload
func NewConfigWithHotReload() (*ConfigWithWatcher, error) {
	config, watcher, err := HotReloadConfig()
	if err != nil {
		return nil, err
	}

	return &ConfigWithWatcher{
		Config:  config,
		Watcher: watcher,
	}, nil
}
