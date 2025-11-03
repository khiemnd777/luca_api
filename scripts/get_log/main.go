package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	// ✅ Parse env flag: --env=dev | --env=production
	env := flag.String("env", "dev", "Environment: dev or production")
	flag.Parse()

	// ✅ Map env -> domain
	var domain string
	switch *env {
	case "production":
		domain = "api.honvang.com"
	default:
		domain = "api-dev.honvang.com"
	}

	config.Init(utils.GetFullPath("config.yaml"))

	url := fmt.Sprintf("https://%s/__log", domain)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("❌ Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Get().Auth.InternalLogToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("❌ Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("❌ Error reading response:", err)
		return
	}

	fmt.Println(string(body))

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
