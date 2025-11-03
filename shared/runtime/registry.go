package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/khiemnd777/andy_api/shared/app"
	"gopkg.in/yaml.v3"
)

type RunningModule struct {
	PID      int       `json:"pid"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	RunAt    time.Time `json:"run_at"`
	External bool      `json:"external"`
	Route    string    `json:"router"`
}

type Registry map[string]RunningModule

var (
	registryPath = "tmp/runtime.json"
	mu           sync.Mutex
)

// LoadRegistry đọc file; nếu chưa có trả map rỗng.
func LoadRegistry() (Registry, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Registry{}, nil
		}
		return nil, err
	}

	reg := Registry{}
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// SaveRegistry ghi đè file.
func SaveRegistry(reg Registry) error {
	data, _ := json.MarshalIndent(reg, "", "  ")
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		return err
	}
	return os.WriteFile(registryPath, data, 0o644)
}

// Register module (gọi trong main.go sau khi biết realPort)
func Register(name, route, host string, port int, external bool) error {
	mu.Lock()
	defer mu.Unlock()

	reg, _ := LoadRegistry()
	reg[name] = RunningModule{
		PID:      os.Getpid(),
		Host:     host,
		Port:     port,
		Route:    route,
		RunAt:    time.Now(),
		External: external,
	}
	return SaveRegistry(reg)
}

// GenerateRegistry duyệt modules/, gán host = "127.0.0.1",
//   - Nếu config.yaml có port>0  ➜ giữ nguyên
//   - Nếu port==0               ➜ auto-allocate
//   - Ghi toàn bộ vào tmp/runtime.json
func GenerateRegistry(modDir string) (Registry, []*app.Reserved, error) {
	reg := Registry{}
	var rs []*app.Reserved

	entries, err := os.ReadDir(modDir)
	if err != nil {
		return nil, nil, err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()

		cfgFile := filepath.Join(modDir, name, "config.yaml")
		raw, routeFromCfg, hostFromCfg, portFromCfg, externalFromCfg, err := loadServerSection(cfgFile)
		if err != nil {
			fmt.Printf("⚠️  skip %s: %v\n", name, err)
			continue
		}

		host := hostFromCfg
		r, portErr := app.EnsurePortAvailable(host, portFromCfg)
		if portErr != nil {
			fmt.Printf("🛑  cannot alloc port for %s\n", name)
			continue
		}

		rs = append(rs, r)

		reg[name] = RunningModule{
			PID:      0,
			Host:     host,
			Port:     r.Port,
			Route:    routeFromCfg,
			RunAt:    time.Now(),
			External: externalFromCfg,
		}

		_ = raw
	}

	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return nil, nil, err
	}
	data, _ := json.MarshalIndent(reg, "", "  ")
	if err := os.WriteFile(registryPath, data, 0o644); err != nil {
		return nil, nil, err
	}
	return reg, rs, nil
}

func loadServerSection(cfgPath string) (
	rawFile []byte, route, host string, port int, external bool, err error,
) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, "", "", 0, false, err
	}

	var raw struct {
		Server struct {
			Host  string `yaml:"host"`
			Port  int    `yaml:"port"`
			Route string `yaml:"route"`
		} `yaml:"server"`
		External bool `yaml:"external"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, "", "", 0, false, err
	}

	return data, raw.Server.Route, raw.Server.Host, raw.Server.Port, raw.External, nil
}
