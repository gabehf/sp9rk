package action

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func CreateApplication(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() != 1 {
			return errors.New("create app must have exactly one argument")
		}
		if !valid(ctx.Args().Get(0)) {
			return errors.New("application name must only contain letters, numbers, dashes and underscores")
		}
		if _, err := os.Stat(AppPath(cfgPath, ctx.Args().Get(0))); err == nil {
			return errors.New("application already exists")
		}
		err := WriteAppFiles(cfgPath, &AppInfo{
			Version:     "1",
			Name:        ctx.Args().Get(0),
			Description: ctx.String("description"),
			Host:        ctx.String("host"),
		})
		if err != nil {
			return err
		}
		fmt.Printf("Created application %s\n", ctx.Args().Get(0))
		return nil
	}
}

func CreateRequest(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 || ctx.NArg() > 1 {
			return errors.New("create req must have exactly one argument")
		}

		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}

		reqName := ctx.Args().Get(0)

		if !valid(app) || !appExists(cfgPath, app) {
			return errors.New("application does not exist")
		}
		if !valid(reqName) {
			return errors.New("request name must only contain letters, numbers, dashes and underscores")
		}

		path := ReqPath(cfgPath, app, reqName)
		if _, err := os.Stat(path); err == nil {
			return errors.New("request already exists")
		}

		if err := WriteRequestFiles(cfgPath, app, &RequestInfo{
			Version:     "1",
			Name:        reqName,
			Description: ctx.String("description"),
			Method:      ctx.String("method"),
			Path:        ctx.String("path"),
			Headers:     ctx.StringSlice("header"),
			Body:        ctx.String("body"),
		}); err != nil {
			return err
		}
		fmt.Printf("Created request %s\n", reqName)
		return nil
	}
}
func EditApplication(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() != 1 {
			return errors.New("edit app must have exactly one argument")
		}
		if !valid(ctx.Args().Get(0)) {
			return errors.New("application name is invalid")
		}
		contents, err := os.ReadFile(AppInfoFilePath(cfgPath, ctx.Args().Get(0)))
		if err != nil {
			return errors.New("application does not exist")
		}
		appinfo := new(AppInfo)
		err = yaml.Unmarshal(contents, appinfo)
		if err != nil {
			return errors.New("failed to read application info")
		}
		if ctx.String("description") != "" {
			appinfo.Description = ctx.String("description")
		}
		if ctx.String("host") != "" {
			appinfo.Host = ctx.String("host")
		}
		err = WriteAppFiles(cfgPath, appinfo)
		if err != nil {
			return err
		}
		return nil
	}
}

func EditRequest(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 || ctx.NArg() > 1 {
			return errors.New("create req must have exactly one argument")
		}

		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}

		reqName := ctx.Args().Get(0)

		if !valid(app) || !appExists(cfgPath, app) {
			return errors.New("application does not exist")
		}
		if !valid(reqName) {
			return errors.New("request name is invalid")
		}

		path := ReqPath(cfgPath, app, reqName)
		contents, err := os.ReadFile(path)
		if err != nil {
			return errors.New("request does not exists")
		}
		reqinfo := new(RequestInfo)
		err = yaml.Unmarshal(contents, reqinfo)
		if err != nil {
			return errors.New("request info is malformed or corrupted")
		}

		if ctx.String("description") != "" {
			reqinfo.Description = ctx.String("description")
		}
		if ctx.String("method") != "" {
			reqinfo.Method = ctx.String("method")
		}
		if ctx.String("path") != "" {
			reqinfo.Path = ctx.String("path")
		}
		if ctx.String("body") != "" {
			reqinfo.Body = ctx.String("body")
		}
		if ctx.StringSlice("header") != nil {
			reqinfo.Headers = ctx.StringSlice("header")
		}

		if err := WriteRequestFiles(cfgPath, app, reqinfo); err != nil {
			return err
		}
		return nil
	}
}

func DeleteApplication(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 || ctx.NArg() > 1 {
			return errors.New("delete app must have exactly one argument")
		}

		app := ctx.Args().Get(0)

		if !valid(app) {
			return errors.New("application name is invalid")
		}

		_path := AppPath(cfgPath, app)

		files, err := os.ReadDir(_path)
		if err != nil {
			return errors.New("application does not exist")
		}
		fmt.Printf(
			"You are about to delete the application %s and %d associated request(s).\nThis action cannot be undone.\n",
			app,
			len(files)-1, // -1 because .appinfo isnt a req
		)
		if ctx.Bool("confirm") || ConfirmPrompt() {
			err := os.RemoveAll(_path)
			if err != nil {
				return err
			}
			fmt.Print("application " + app + " has been deleted")
			return nil
		}
		fmt.Println("delete aborted")
		return nil
	}
}
func DeleteRequest(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {

		if ctx.NArg() < 1 || ctx.NArg() > 1 {
			return errors.New("delete request must have exactly one argument")
		}

		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}

		reqName := ctx.Args().Get(0)

		if !valid(reqName) {
			return errors.New("request name is invalid")
		}
		if !valid(app) || !appExists(cfgPath, app) {
			return errors.New("application does not exist")
		}

		_path := ReqPath(cfgPath, app, reqName)

		if _, err := os.Stat(_path); err != nil {
			return errors.New("request does not exist")
		}

		fmt.Printf(
			"You are about to delete the request %s in application %s.\nThis action cannot be undone.\n",
			reqName,
			app,
		)
		if ctx.Bool("confirm") || ConfirmPrompt() {
			err := os.Remove(_path)
			if err != nil {
				return err
			}
			fmt.Print("request " + reqName + " has been deleted")
			return nil
		}
		fmt.Println("delete aborted")
		return nil
	}
}

func Switch(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("expected argument")
		}
		if ctx.NArg() > 1 {
			return errors.New("expected exactly one argument")
		}
		app := ctx.Args().Get(0)
		if !valid(app) {
			return errors.New("application does not exist")
		}
		if _, err := os.Stat(AppPath(cfgPath, app)); err != nil {
			return errors.New("application does not exist")
		}
		err := os.WriteFile(path.Join(cfgPath, "current_app"), []byte(ctx.Args().First()), 0700)
		if err != nil {
			return err
		}
		fmt.Print(app)
		return nil
	}
}

func ListApplications(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		_path := AppPath(cfgPath, "")
		dirs, err := os.ReadDir(_path)
		if err != nil {
			return err
		}
		var apps string
		for _, dir := range dirs {
			appinfo := new(AppInfo)
			contents, err := os.ReadFile(AppInfoFilePath(cfgPath, dir.Name()))
			if err != nil {
				return errors.New("failed to read application info")
			}
			err = yaml.Unmarshal(contents, appinfo)
			if err != nil {
				return errors.New(".appinfo is malformed or corrupted")
			}
			if appinfo.Description == "" {
				apps += appinfo.Name + "\n"
			} else {
				apps += appinfo.Name + ": " + appinfo.Description + "\n"
			}
		}
		fmt.Print(apps)
		return nil
	}
}

func ListRequests(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}
		_path := AppPath(cfgPath, app)
		files, err := os.ReadDir(_path)
		if err != nil {
			return err
		}
		reqs := ""
		for _, file := range files {
			if file.Name() == ".appinfo" {
				continue
			}
			reqinfo := new(RequestInfo)
			contents, err := os.ReadFile(path.Join(AppPath(cfgPath, app), file.Name()))
			if err != nil {
				return errors.New("failed to read request info")
			}
			err = yaml.Unmarshal(contents, reqinfo)
			if err != nil {
				return errors.New("request file is malformed or corrupted")
			}
			if reqinfo.Description == "" {
				reqs += reqinfo.Name + "\n"
			} else {
				reqs += reqinfo.Name + ": " + reqinfo.Description + "\n"
			}
		}
		fmt.Print(reqs)
		return nil
	}
}
func ListAll(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		apps, err := os.ReadDir(AppPath(cfgPath, ""))
		if err != nil {
			return err
		}
		output := ""
		if len(apps) < 1 {
			fmt.Println("No applications or requests")
			return nil
		}
		for _, app := range apps {
			reqfiles, err := os.ReadDir(AppPath(cfgPath, app.Name()))
			if err != nil {
				return err
			}
			if len(reqfiles) <= 1 {
				output += app.Name() + "\n"
			} else {
				output += app.Name() + ":\n"
			}
			for _, reqfile := range reqfiles {
				if reqfile.Name() == ".appinfo" {
					continue
				}
				reqinfo := new(RequestInfo)
				contents, err := os.ReadFile(path.Join(AppPath(cfgPath, app.Name()), reqfile.Name()))
				if err != nil {
					return errors.New("failed to read request info")
				}
				err = yaml.Unmarshal(contents, reqinfo)
				if err != nil {
					return errors.New("request file is malformed or corrupted")
				}
				if reqinfo.Description == "" {
					output += "\t- " + reqinfo.Name + "\n"
				} else {
					output += "\t- " + reqinfo.Name + ": " + reqinfo.Description + "\n"
				}
			}
		}
		fmt.Print(output)
		return nil
	}
}
func InfoApplication(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("expected argument")
		}
		if ctx.NArg() > 1 {
			return errors.New("expected exactly one argument")
		}
		if !valid(ctx.Args().Get(0)) {
			return errors.New("application name is invalid")
		}
		contents, err := os.ReadFile(AppInfoFilePath(cfgPath, ctx.Args().Get(0)))
		if err != nil {
			return errors.New("failed to read app info")
		}
		appinfo := new(AppInfo)
		err = yaml.Unmarshal(contents, appinfo)
		if err != nil {
			return errors.New("failed to read app info")
		}
		fmt.Printf("%s:\n\tDescription: %s\n\tHost: %s\n", appinfo.Name, appinfo.Description, appinfo.Host)
		return nil
	}
}
func InfoRequest(cfgPath string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("expected argument")
		}
		if ctx.NArg() > 1 {
			return errors.New("expected exactly one argument")
		}
		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}
		if !valid(app) {
			return errors.New("application name is invalid")
		}
		if !valid(ctx.Args().Get(0)) {
			return errors.New("request name is invalid")
		}
		contents, err := os.ReadFile(ReqPath(cfgPath, app, ctx.Args().Get(0)))
		if err != nil {
			return errors.New("failed to read request info")
		}
		fileLines := strings.Split(string(contents), "\n")
		fmt.Println(ctx.Args().Get(0) + ":\n\t" + strings.Join(fileLines[2:len(fileLines)-1], "\n\t"))
		return nil
	}
}

type VerboseCallResponse struct {
	Request      string            `yaml:"Request"`
	RequestBody  string            `yaml:"RequestBody"`
	Headers      map[string]string `yaml:"Headers"`
	Latency      string            `yaml:"Latency"`
	Status       string            `yaml:"Status"`
	ResponseBody string            `yaml:"ResponseBody"`
}

func Call(cfgPath string, httpClient http.Client) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("expected argument")
		}
		if ctx.NArg() > 1 {
			return errors.New("expected exactly one argument")
		}
		app, err := getApp(cfgPath, ctx)
		if err != nil {
			return err
		}
		if !valid(app) {
			return errors.New("application name is invalid")
		}
		if !valid(ctx.Args().Get(0)) {
			return errors.New("request name is invalid")
		}
		// retrieve app information
		contents, err := os.ReadFile(AppInfoFilePath(cfgPath, app))
		if err != nil {
			return errors.New("failed to retrieve app information")
		}
		appinfo := new(AppInfo)
		err = yaml.Unmarshal(contents, appinfo)
		if err != nil {
			return errors.New("appinfo file is malformed or corrupted")
		}
		// retrieve request information
		contents, err = os.ReadFile(ReqPath(cfgPath, app, ctx.Args().Get(0)))
		if err != nil {
			return errors.New("failed to retrieve request information")
		}
		reqinfo := new(RequestInfo)
		err = yaml.Unmarshal(contents, reqinfo)
		if err != nil {
			return errors.New("request file is malformed or corrupted")
		}
		body := bytes.NewBuffer([]byte(reqinfo.Body))
		req, err := http.NewRequest(reqinfo.Method, appinfo.Host+reqinfo.Path, body)
		if err != nil {
			return errors.New("failed to create web request")
		}
		for _, header := range reqinfo.Headers {
			h := strings.Split(header, ": ")
			if len(h) < 2 {
				return errors.New("malformed header(s)")
			}
			req.Header.Add(h[0], h[1])
		}
		if ctx.Bool("no-redirect") {
			httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
		}
		t1 := time.Now()
		resp, err := httpClient.Do(req)
		t2 := time.Now()
		if err != nil {
			return errors.New("failed to send web request")
		}
		defer resp.Body.Close()
		if ctx.Bool("fail") && resp.StatusCode >= 400 {
			return errors.New("")
		}
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New("failed to read response body")
		}
		if !ctx.Bool("verbose") {
			fmt.Print(string(respBody) + "\n")
		} else if req.Method == "GET" {
			var out VerboseCallResponse
			out.Request = req.Method + " " + req.URL.String()
			out.RequestBody = reqinfo.Body
			out.Headers = make(map[string]string)
			for _, header := range reqinfo.Headers {
				k := strings.Split(header, ": ")[0]
				v := strings.Split(header, ": ")[1]
				out.Headers[k] = v
			}
			out.Latency = fmt.Sprintf("%v", t2.Sub(t1))
			out.Status = resp.Status
			out.ResponseBody = string(respBody)
			o, err := yaml.Marshal(out)
			if err != nil {
				return errors.New("failed to generate command output")
			}
			fmt.Print(string(o))
		}
		return nil
	}
}
