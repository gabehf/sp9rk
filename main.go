package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gabehf/sp9rk/app"
)

func main() {
	var cfgPath string
	if os.Getenv("SP9RK_CONFIG_PATH") != "" {
		cfgPath = os.Getenv("SP9RK_CONFIG_PATH")
	} else {
		homedir, err := os.UserConfigDir()
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to retrieve user config directory")
			os.Exit(1)
		}
		cfgPath = path.Join(homedir, "sp9rk")
	}
	// init app
	app := app.New(cfgPath, http.Client{})

	if err := app.Run(os.Args); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
