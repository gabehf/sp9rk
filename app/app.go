package app

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gabehf/sp9rk/action"
	"github.com/urfave/cli/v2"
)

func New(cfgPath string, httpClient http.Client) *cli.App {

	// prepare necessary directories
	if _, err := os.Stat(cfgPath); err != nil {
		err := os.MkdirAll(cfgPath, 0700)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to create configuration directory")
			os.Exit(1)
		}
	}
	if _, err := os.Stat(action.AppPath(cfgPath, "")); err != nil {
		err := os.MkdirAll(action.AppPath(cfgPath, ""), 0700)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to create apps directory")
			os.Exit(1)
		}
	}
	// prepare config files
	if _, err := os.Stat(path.Join(cfgPath, "current_app")); err != nil {
		err := os.WriteFile(path.Join(cfgPath, "current_app"), []byte(""), 0700)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to create required configuration file")
			os.Exit(1)
		}
	}

	debugFlag := "debug"
	confirmFlag := "confirm"
	appFlag := []string{"app", "a"}
	descriptionFlag := []string{"description", "d"}
	hostFlag := []string{"host", "u"}
	methodFlag := []string{"method", "X"}
	pathFlag := []string{"path", "p"}
	bodyFlag := []string{"body", "b"}
	headerFlag := []string{"header", "H"}
	verboseFlag := []string{"verbose", "v"}
	noRedirectFlag := []string{"no-redirect", "n"}
	failFlag := []string{"fail", "f"}

	return &cli.App{
		Name:    "sp9rk",
		Usage:   "Automate your API calls in the command line",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  debugFlag,
				Usage: "enable debug output",
			},
		},
		Action: func(*cli.Context) error {
			fmt.Print(`           ___       _    
 ___ _ __ / _ \ _ __| | __
/ __| '_ \ (_) | '__| |/ /
\__ \ |_) \__, | |  |   < 
|___/ .__/  /_/|_|  |_|\_\
    |_|                   

Automate your API calls in the command line!
Type 'sp9rk help' for more information.`)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create applications or requests",
				Subcommands: []*cli.Command{
					{
						Name:  "app",
						Usage: "create an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    descriptionFlag[0],
								Aliases: descriptionFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    hostFlag[0],
								Aliases: hostFlag[1:],
								Usage:   "specify the application's host address",
								Value:   "http://localhost",
							},
						},
						Action: action.CreateApplication(cfgPath),
					},
					{
						Name:    "request",
						Aliases: []string{"req"},
						Usage:   "create a request within an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    appFlag[0],
								Aliases: appFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    methodFlag[0],
								Aliases: methodFlag[1:],
								Value:   "GET",
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    pathFlag[0],
								Aliases: pathFlag[1:],
								Value:   "/",
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    descriptionFlag[0],
								Aliases: descriptionFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    bodyFlag[0],
								Aliases: bodyFlag[1:],
								Usage:   "",
							},
							&cli.StringSliceFlag{
								Name:    headerFlag[0],
								Aliases: headerFlag[1:],
								Usage:   "",
							},
						},
						Action: action.CreateRequest(cfgPath),
					},
				},
			},
			{
				Name:  "edit",
				Usage: "edit an existing application or request",
				Subcommands: []*cli.Command{
					{
						Name:  "app",
						Usage: "edit an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    descriptionFlag[0],
								Aliases: descriptionFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    hostFlag[0],
								Aliases: hostFlag[1:],
								Usage:   "specify the application's host address",
								Value:   "http://localhost",
							},
						},
						Action: action.EditApplication(cfgPath),
					},
					{
						Name:    "request",
						Aliases: []string{"req"},
						Usage:   "edit a request within an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    appFlag[0],
								Aliases: appFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    methodFlag[0],
								Aliases: methodFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    pathFlag[0],
								Aliases: pathFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    descriptionFlag[0],
								Aliases: descriptionFlag[1:],
								Usage:   "",
							},
							&cli.StringFlag{
								Name:    bodyFlag[0],
								Aliases: bodyFlag[1:],
								Usage:   "",
							},
							&cli.StringSliceFlag{
								Name:    headerFlag[0],
								Aliases: headerFlag[1:],
								Usage:   "",
							},
						},
						Action: action.EditRequest(cfgPath),
					},
				},
			},
			{
				Name:  "delete",
				Usage: "delete applications or requests",
				Subcommands: []*cli.Command{
					{
						Name:  "app",
						Usage: "delete an application and all associated requests",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  confirmFlag,
								Usage: "skips the confirmation request and immediately deletes the resource",
							},
						},
						Action: action.DeleteApplication(cfgPath),
					},
					{
						Name:    "request",
						Aliases: []string{"req"},
						Usage:   "delete a request within an application",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  confirmFlag,
								Usage: "skips the confirmation request and immediately deletes the resource",
							},
							&cli.StringFlag{
								Name:    appFlag[0],
								Aliases: appFlag[1:],
								Usage:   "specify an application",
							},
						},
						Action: action.DeleteRequest(cfgPath),
					},
				},
			},
			{
				Name:   "list",
				Usage:  "list out applications or requests",
				Action: action.ListAll(cfgPath),
				Subcommands: []*cli.Command{
					{
						Name:   "app",
						Usage:  "list saved applications",
						Action: action.ListApplications(cfgPath),
					},
					{
						Name:    "request",
						Aliases: []string{"req"},
						Usage:   "list requests within an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    appFlag[0],
								Aliases: appFlag[1:],
								Usage:   "specify an application",
							},
						},
						Action: action.ListRequests(cfgPath),
					},
				},
			},
			{
				Name:  "info",
				Usage: "info of an application or request",
				Subcommands: []*cli.Command{
					{
						Name:   "app",
						Usage:  "list saved applications",
						Action: action.InfoApplication(cfgPath),
					},
					{
						Name:    "request",
						Aliases: []string{"req"},
						Usage:   "list requests within an application",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    appFlag[0],
								Aliases: appFlag[1:],
								Usage:   "specify an application",
							},
						},
						Action: action.InfoRequest(cfgPath),
					},
				},
			},
			{
				Name:   "switch",
				Usage:  "set your current app",
				Action: action.Switch(cfgPath),
			},
			{
				Name:  "call",
				Usage: "make a request",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    appFlag[0],
						Aliases: appFlag[1:],
						Usage:   "specify an application",
					},
					&cli.BoolFlag{
						Name:    verboseFlag[0],
						Aliases: verboseFlag[1:],
						Usage:   "enable verbose output",
					},
					&cli.BoolFlag{
						Name:    failFlag[0],
						Aliases: failFlag[1:],
						Usage:   "fail silently",
					},
					&cli.BoolFlag{
						Name:    noRedirectFlag[0],
						Aliases: noRedirectFlag[1:],
						Usage:   "follow redirects",
					},
				},
				Action: action.Call(cfgPath, httpClient),
			},
		},
	}
}
