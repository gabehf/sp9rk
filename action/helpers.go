package action

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

// If the --app flag is set, returns that app. Otherwise, returns the default app.
// If no default exists, returns an error.
func getApp(cfgPath string, ctx *cli.Context) (string, error) {
	var app string
	if ctx.String("app") == "" {
		app = currentApp(cfgPath)
		if app == "" {
			return "", errors.New("application does not exist")
		}
	} else {
		app = ctx.String("app")
	}
	return app, nil
}

// "" if no current app is set
func currentApp(cfgPath string) string {
	if _, err := os.Stat(path.Join(cfgPath, "current_app")); err != nil {
		return ""
	}
	app, err := os.ReadFile(path.Join(cfgPath, "current_app"))
	if err != nil {
		return ""
	}
	return string(app)
}

// TRUE if string is alphanumeric with - or _
func valid(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]*$`).MatchString(name)
}

// TRUE if app folder exists
func appExists(cfgPath, app string) bool {
	_, err := os.Stat(AppPath(cfgPath, app))
	return err == nil
}

func ConfirmPrompt() bool {
	fmt.Print("Are you sure? [y/N]: ")
	r := bufio.NewReader(os.Stdin)
	res, err := r.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	// if empty, default to N
	if len(res) < 2 {
		return false
	}
	return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
}

func AppPath(cfgPath, app string) string {
	return path.Join(cfgPath, "apps", app)
}

func AppInfoFilePath(cfgPath, app string) string {
	return path.Join(AppPath(cfgPath, app), ".appinfo")
}

func ReqPath(cfgPath, app, req string) string {
	return path.Join(cfgPath, "apps", app, req+".yml")
}
