package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/khiemnd777/andy_api/scripts/module_runner/runner"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❗ Usage: start|start-all|stop|stop-all|sync <module>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start", "stop":
		if len(os.Args) < 3 {
			fmt.Printf("❗ Usage: %s <module>\n", command)
			os.Exit(1)
		}
		module := os.Args[2]

		checkErr(runner.SyncRunningModules())
		if command == "start" {
			checkErr(runner.StartModule(module))
		} else {
			checkErr(runner.StopModule(module))
		}
		checkErr(runner.SyncRunningModules())
	case "start-all":
		checkErr(runner.SyncRunningModules())
		dscvrModules, err := utils.DiscoverAllModules()
		checkErr(err)
		// checkErr(runner.StartModulesInBatch([]string{"auditlog", "auth", "auth_facebook", "auth_google", "permission"}))
		checkErr(runner.StartModulesInBatch(dscvrModules))
		time.Sleep(1 * time.Second)
		checkErr(runner.SyncRunningModules())
	case "stop-all":
		checkErr(runner.SyncRunningModules())
		checkErr(runner.StopAllModules())
		checkErr(runner.SyncRunningModules())
	case "sync":
		checkErr(runner.SyncRunningModules())
	case "status":
		checkErr(runner.ShowStatus())
	default:
		fmt.Println("❓ Unknown command. Use: start | stop | sync")
		os.Exit(1)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
