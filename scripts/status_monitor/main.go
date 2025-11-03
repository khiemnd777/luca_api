package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
	"github.com/khiemnd777/andy_api/shared/utils"
	"gopkg.in/yaml.v3"
)

type ModuleStatus struct {
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
	Message   string    `json:"message,omitempty"`
}

const (
	statusPrefix     = "module_status:"
	defaultTimeout   = 1 * time.Minute
	warningTimeout   = 5 * time.Minute
	configFileName   = "config.yaml"
	moduleFolderPath = "modules"
	redisInstance    = "status"
)

func main() {
	logger.Init()

	cfg, err := utils.LoadConfig[config.AppConfig](utils.GetFullPath("config.yaml"))
	if err != nil {
		fmt.Printf("❌ Cannot load config.yaml: %v\n", err)
		return
	}

	redis.InitFromConfig(cfg.Redis)

	printAsJSON := len(os.Args) > 1 && os.Args[1] == "--json"

	fmt.Println("📡 Scanning modules...")
	entries, err := os.ReadDir(moduleFolderPath)
	if err != nil {
		fmt.Printf("❌ Cannot read modules folder: %v\n", err)
		return
	}

	var results []ModuleStatus

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleName := entry.Name()
		modulePath := filepath.Join(moduleFolderPath, moduleName)
		configPath := filepath.Join(modulePath, configFileName)

		// Read config.yaml
		cfg := struct {
			Server struct {
				Port int `yaml:"port"`
			} `yaml:"server"`
		}{}

		port := 0
		if data, err := os.ReadFile(configPath); err == nil {
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				port = cfg.Server.Port
			}
		}

		status := getStatusFromRedis(moduleName)
		status.Name = moduleName
		status.Port = port
		results = append(results, status)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	if printAsJSON {
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	printTable(results)
}

func getStatusFromRedis(module string) ModuleStatus {
	rdb := redis.GetInstance(redisInstance)
	if rdb == nil {
		return ModuleStatus{
			Status:  "ERROR",
			Message: "redis instance not found",
		}
	}

	ctx := context.Background()
	key := statusPrefix + module

	raw, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return ModuleStatus{
			Status:  "MISSING",
			Message: "status not found",
		}
	}

	var ms ModuleStatus
	if err := json.Unmarshal([]byte(raw), &ms); err != nil {
		return ModuleStatus{
			Status:  "ERROR",
			Message: "invalid status format",
		}
	}

	if time.Since(ms.UpdatedAt) > defaultTimeout {
		ms.Status = "STOPPED"
	}

	return ms
}

func printTable(statuses []ModuleStatus) {
	fmt.Println()
	fmt.Printf("%-15s %-6s %-10s %-20s %-s\n", "Module", "Port", "Status", "Updated At", "Message")
	fmt.Println(strings.Repeat("-", 80))

	for _, s := range statuses {
		statusText := s.Status
		switch s.Status {
		case "RUNNING":
			statusText = color.HiGreenString("RUNNING")
		case "FAILED":
			statusText = color.HiRedString("FAILED")
		case "STOPPED":
			statusText = color.YellowString("STOPPED")
		case "MISSING":
			statusText = color.YellowString("MISSING")
		default:
			statusText = color.CyanString(s.Status)
		}

		timeStr := "—"
		if !s.UpdatedAt.IsZero() {
			timeStr = s.UpdatedAt.Format("2006-01-02 15:04:05")
		}

		if (s.Status == "FAILED" || s.Status == "STOPPED") && time.Since(s.UpdatedAt) > warningTimeout {
			s.Message += " ⚠️ WARNING: stale module!"
		}

		fmt.Printf("%-15s %-6d %-10s %-20s %-s\n", s.Name, s.Port, statusText, timeStr, s.Message)
	}
}
