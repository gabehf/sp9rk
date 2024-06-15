package action_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gabehf/sp9rk/action"
	"github.com/gabehf/sp9rk/app"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// pretty much all the other tests rely on this test working...
// i know its bad practice but I dont want to manually create every mock app and request files
// manually every time i make the test.
//
// Basically, if this test fails, fix it FIRST.
func TestWriteFiles(t *testing.T) {
	cfgPath := path.Join("TestActionWriteFiles", ".sp9rk", "tests")
	err := action.WriteAppFiles(cfgPath, &action.AppInfo{
		Version:     "1",
		Name:        "TestApp",
		Description: "test app",
		Host:        "localhost",
	})
	assert.NoError(t, err, "FIX FIRST: failed to write app files")
	p := action.AppPath(cfgPath, "TestApp")
	_, err = os.Stat(p)
	assert.NoError(t, err, "failed to create app directory")
	_, err = os.Stat(path.Join(p, ".appinfo"))
	assert.NoError(t, err, "failed to create .appinfo file")
	err = action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Version:     "1",
		Name:        "MyReq",
		Description: "my request",
	})
	assert.NoError(t, err, "FIX FIRST: failed to write request file")
	_, err = os.Stat(path.Join(p, "MyReq.yml"))
	assert.NoError(t, err, "failed to stat request file")
	os.RemoveAll("TestActionWriteFiles")
}

func TestActionCreateApplication(t *testing.T) {
	cfgPath := path.Join("TestActionCreateApplication", ".sp9rk", "tests")
	app := app.New(cfgPath, http.Client{})
	assert.Error(
		t,
		RunWithArgs(app, "create", "app"),
		"create app should fail with no args",
	)
	assert.Error(
		t,
		RunWithArgs(app, "create", "app", "-u", "hostname123", "../malicious/path"),
		"create app should fail invalid name",
	)
	assert.Error(
		t,
		RunWithArgs(app, "create", "app", "-u", "hostname123", "TestApp", "secondArgument"),
		"create app should fail with >1 args",
	)
	assert.NoError(
		t,
		RunWithArgs(app, "create", "app", "-u", "hostname123", "TestApp"),
		"create app should succeed with 1 args",
	)
	assert.Error(
		t,
		RunWithArgs(app, "create", "app", "-u", "hostname123", "TestApp"),
		"create app should fail when app already exists",
	)
	// cleanup
	os.RemoveAll("TestActionCreateApplication")
}
func TestActionCreateRequest(t *testing.T) {
	cfgPath := path.Join("TestActionCreateRequest", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestAppTwo",
	})
	// tests
	app := app.New(cfgPath, http.Client{})

	assert.Error(t, RunWithArgs(app, "create", "req"), "create req should fail with no args")
	assert.Error(t, RunWithArgs(app, "create", "req", "TestReq"), "create req should fail with no default app")
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	assert.NoError(t, RunWithArgs(app, "create", "req", "TestReq"), "create req should succeed with default app")
	assert.Error(t, RunWithArgs(app, "create", "req", "TestReq", "SecondArg"), "create req should fail with >1 args")
	assert.NoError(t, RunWithArgs(app, "create", "req", "-a", "TestAppTwo", "TestReq"), "create req should succeed with specified app")
	assert.Error(t, RunWithArgs(app, "create", "req", "-a", "TestAppThree", "TestReq"), "create req should fail with unknown app")
	assert.Error(t, RunWithArgs(app, "create", "req", "../malicious/path"), "create req should fail invalid name")
	assert.Error(t, RunWithArgs(app, "create", "req", "-a", "TestAppTwo", "TestReq"), "create req should fail if it already exists")
	contents, _ := os.ReadFile(path.Join(action.AppPath(cfgPath, "TestApp"), "TestReq.yml"))
	reqinfo := new(action.RequestInfo)
	err := yaml.Unmarshal(contents, reqinfo)
	assert.NoError(t, err, "contents should be YAML")
	assert.EqualValues(t, reqinfo.Name, "TestReq")
	assert.EqualValues(t, reqinfo.Description, "")
	assert.EqualValues(t, reqinfo.Method, "GET")
	assert.EqualValues(t, reqinfo.Path, "/")
	assert.EqualValues(t, reqinfo.Body, "")
	assert.Empty(t, reqinfo.Headers)

	// flag tests
	assert.NoError(
		t,
		RunWithArgs(app, "create", "req",
			"-d", "description",
			"-X", "POST",
			"-p", "/path",
			"-b", "request body",
			"-H", "X-Header-One: 1",
			"-H", "X-Header-Two: 2",
			"TestReqTwo",
		),
		"create req should succeed with flags",
	)
	contents, _ = os.ReadFile(path.Join(action.AppPath(cfgPath, "TestApp"), "TestReqTwo.yml"))
	reqinfo = new(action.RequestInfo)
	err = yaml.Unmarshal(contents, reqinfo)
	assert.NoError(t, err, "contents should be YAML")
	if reqinfo.Headers == nil || len(reqinfo.Headers) < 2 {
		assert.FailNow(t, "file contents were not parsed successfully")
	}
	assert.EqualValues(t, reqinfo.Name, "TestReqTwo")
	assert.EqualValues(t, reqinfo.Description, "description")
	assert.EqualValues(t, reqinfo.Method, "POST")
	assert.EqualValues(t, reqinfo.Path, "/path")
	assert.EqualValues(t, reqinfo.Body, "request body")
	assert.EqualValues(t, reqinfo.Headers[0], "X-Header-One: 1")
	assert.EqualValues(t, reqinfo.Headers[1], "X-Header-Two: 2")
	// cleanup
	os.RemoveAll("TestActionCreateRequest")
}
func TestActionSwitch(t *testing.T) {
	cfgPath := path.Join("TestActionSwitch", ".sp9rk", "tests")
	app := app.New(cfgPath, http.Client{})
	assert.Error(t, RunWithArgs(app, "switch"), "switch should fail with no args")
	assert.Error(t, RunWithArgs(app, "switch", "TestApp"), "switch should fail with no apps existing")
	// simulate apps being created and one set as default
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestAppTwo",
	})
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	// tests
	assert.Error(t, RunWithArgs(app, "switch"), "switch should fail with no args")
	output, err := captureOutput(RunWithArgs, app, "switch", "TestApp")
	assert.NoError(t, err, "switch should NOT fail with existing app")
	assert.Equal(t, "TestApp", output, "switch output should be equal to new app name on success")
	assert.Error(t, RunWithArgs(app, "switch", "../malicious/path"), "switch should fail invalid name")
	assert.Error(t, RunWithArgs(app, "switch", "TestApp", "somethingElse"), "switch should fail with >1 arg")
	assert.Error(t, RunWithArgs(app, "switch", "NotRealApp"), "switch should fail with unknown app")
	os.RemoveAll("TestActionSwitch")
}
func TestActionDeleteApp(t *testing.T) {
	cfgPath := path.Join("TestActionDeleteApp", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "TestApp",
		Description: "this is a very cool app that does cool stuff",
		Host:        "http://localhost:3000",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "TestAppTwo",
		Description: "this is another very cool app that does cool stuff",
		Host:        "http://localhost:3001",
	})
	// tests
	app := app.New(cfgPath, http.Client{})
	assert.Error(t, RunWithArgs(app, "delete", "app"), "delete app should fail with no args")
	assert.NoError(t, RunWithArgs(app, "delete", "app", "--confirm", "TestApp"), "delete app should NOT fail with existing app")
	assert.Error(t, RunWithArgs(app, "delete", "app", "../malicious/path"), "delete app should fail invalid name")
	assert.Error(t, RunWithArgs(app, "delete", "app", "TestAppTwo", "somethingElse"), "delete app should fail with >1 arg")
	assert.Error(t, RunWithArgs(app, "delete", "app", "NotRealApp"), "delete app should fail with unknown app")
	assert.NoError(t, RunWithArgs(app, "delete", "app", "--confirm", "TestAppTwo"), "delete app should NOT fail with existing app")
	assert.Error(t, RunWithArgs(app, "delete", "app", "TestAppTwo"), "delete app should fail with deleted app")
	os.RemoveAll("TestActionDeleteApp")
}
func TestActionDeleteRequest(t *testing.T) {
	cfgPath := path.Join("TestActionDeleteRequest", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "TestApp",
		Description: "this is a very cool app that does cool stuff",
		Host:        "http://localhost:3000",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "TestAppTwo",
		Description: "this is a very cool app that does cool stuff",
		Host:        "http://localhost:3001",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Version: "1",
		Name:    "MyReq",
	})
	action.WriteRequestFiles(cfgPath, "TestAppTwo", &action.RequestInfo{
		Version: "1",
		Name:    "AnotherReq",
	})
	// tests
	app := app.New(cfgPath, http.Client{})
	assert.Error(t, RunWithArgs(app, "delete", "req"), "delete req should fail with no args")
	assert.Error(t, RunWithArgs(app, "delete", "req", "--confirm", "MyReq"), "delete req should fail with no default app")
	assert.NoError(t, RunWithArgs(app, "delete", "req", "--confirm", "-a", "TestApp", "MyReq"), "delete req should NOT fail with existing req")
	assert.Error(t, RunWithArgs(app, "delete", "req", "-a", "TestApp", "MyReq"), "delete req should fail with deleted req")
	assert.Error(t, RunWithArgs(app, "delete", "req", "-a", "TestApp", "../malicious/path"), "delete req should fail invalid name")
	// create default app
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestAppTwo"), 0700)
	assert.Error(t, RunWithArgs(app, "delete", "req", "AnotherReq", "somethingElse"), "delete req should fail with >1 arg")
	assert.Error(t, RunWithArgs(app, "delete", "req", "NotRealApp"), "delete req should fail with unknown req")
	assert.NoError(t, RunWithArgs(app, "delete", "req", "--confirm", "AnotherReq"), "delete req should NOT fail with default app and existing req")
	assert.Error(t, RunWithArgs(app, "delete", "req", "AnotherReq"), "delete req should fail with deleted req")
	os.RemoveAll("TestActionDeleteRequest")
	// tests
}

func TestActionEditApp(t *testing.T) {
	cfgPath := path.Join("TestActionEditApp", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	// tests
	app := app.New(cfgPath, http.Client{})
	assert.NoError(t,
		RunWithArgs(app, "edit", "app", "-d", "new description", "-u", "http://123.456.789", "TestApp"),
		"edit app should not fail with 1 arg and flags",
	)
	contents, err := os.ReadFile(action.AppInfoFilePath(cfgPath, "TestApp"))
	assert.NoError(t, err, "new .appinfo file should exist")
	appinfo := new(action.AppInfo)
	err = yaml.Unmarshal(contents, appinfo)
	assert.NoError(t, err, "new .appinfo file should be correctly formed")
	assert.EqualValues(t, "new description", appinfo.Description, "description should be updated")
	assert.EqualValues(t, "http://123.456.789", appinfo.Host, "host should be updated")
	assert.Error(t, RunWithArgs(app, "edit", "app"), "edit app should fail with no args")
	assert.NoError(t, RunWithArgs(app, "edit", "app", "TestApp"), "edit app should NOT fail with no flags")
	os.RemoveAll("TestActionEditApp")
}

func TestActionEditRequest(t *testing.T) {
	cfgPath := path.Join("TestActionEditRequest", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name: "Req",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "ReqTwo",
		Path:        "/path",
		Description: "hello",
		Body:        "request body",
		Method:      "POST",
		Headers:     []string{"Header: One"},
	})
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	// tests
	app := app.New(cfgPath, http.Client{})
	assert.NoError(t,
		RunWithArgs(app, "edit", "req",
			"-a", "TestApp",
			"-d", "new description",
			"-X", "POST",
			"-b", "new body",
			"-H", "X-Header: ABC",
			"-p", "/new-path",
			"Req",
		),
		"edit app should not fail with 1 arg and flags",
	)
	contents, err := os.ReadFile(path.Join(action.AppPath(cfgPath, "TestApp"), "Req.yml"))
	assert.NoError(t, err, "new request file should exist")
	reqinfo := new(action.RequestInfo)
	err = yaml.Unmarshal(contents, reqinfo)
	assert.NoError(t, err, "new request file should be correctly formed")
	assert.EqualValues(t, "new description", reqinfo.Description, "description should be updated")
	assert.EqualValues(t, "POST", reqinfo.Method, "method should be updated")
	assert.EqualValues(t, "new body", reqinfo.Body, "body should be updated")
	assert.EqualValues(t, "X-Header: ABC", reqinfo.Headers[0], "header(s) should be updated")
	assert.EqualValues(t, "/new-path", reqinfo.Path, "path should be updated")
	assert.Error(t, RunWithArgs(app, "edit", "req"), "edit req should fail with no args")
	assert.NoError(t, RunWithArgs(app, "edit", "req", "ReqTwo"), "edit req should NOT fail with no flags")
	assert.NoError(t, RunWithArgs(app, "edit", "req", "-d", "new description", "ReqTwo"), "edit req should NOT fail when updating description")
	contents, _ = os.ReadFile(path.Join(action.AppPath(cfgPath, "TestApp"), "ReqTwo.yml"))
	reqinfo = new(action.RequestInfo)
	yaml.Unmarshal(contents, reqinfo)
	assert.EqualValues(t, "Header: One", reqinfo.Headers[0], "header(s) should NOT be overwritten when not updated")
	assert.EqualValues(t, "/path", reqinfo.Path, "path should NOT be overwritten when not updated")
	assert.EqualValues(t, "request body", reqinfo.Body, "header(s) should NOT be overwritten when not updated")
	assert.EqualValues(t, "POST", reqinfo.Method, "method should NOT be overwritten when not updated")
	os.RemoveAll("TestActionEditRequest")
}

func TestActionListApps(t *testing.T) {
	cfgPath := path.Join("TestActionListApps", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "AnotherApp",
		Description: "another one",
	})
	// tests
	expected := "AnotherApp: another one\nTestApp\n"
	app := app.New(cfgPath, http.Client{})
	out, err := captureOutput(RunWithArgs, app, "list", "app")
	assert.NoError(t, err, "list app should NOT generate an error")
	assert.EqualValues(t, expected, out, "list app should be alphabetical")
	os.RemoveAll("TestActionListApps")
}

func TestActionListRequests(t *testing.T) {
	cfgPath := path.Join("TestActionListRequests", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestAppTwo",
	})
	action.WriteRequestFiles(cfgPath, "TestAppTwo", &action.RequestInfo{
		Name: "Req",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Description: "my request",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name: "AnotherReq",
	})
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	// tests
	expected := "AnotherReq\nMyReq: my request\n"
	app := app.New(cfgPath, http.Client{})
	out, err := captureOutput(RunWithArgs, app, "list", "req")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list req should succeed with default app")
	expected = "Req\n"
	out, err = captureOutput(RunWithArgs, app, "list", "req", "-a", "TestAppTwo")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list req should succeed with specific app")
	os.RemoveAll("TestActionListRequests")
}

func TestActionListAll(t *testing.T) {
	cfgPath := path.Join("TestActionListAll", ".sp9rk", "tests")
	// simulate apps being created

	app := app.New(cfgPath, http.Client{})

	expected := "No applications or requests\n"
	out, err := captureOutput(RunWithArgs, app, "list")
	assert.NoError(t, err, "list all with nothing loaded should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output")
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestAppTwo",
	})
	expected = "TestAppTwo\n"
	out, err = captureOutput(RunWithArgs, app, "list")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output")

	action.WriteRequestFiles(cfgPath, "TestAppTwo", &action.RequestInfo{
		Name: "Req",
	})
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Description: "my request",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name: "AnotherReq",
	})
	expected = `TestApp:
	- AnotherReq
	- MyReq: my request
TestAppTwo:
	- Req
`
	out, err = captureOutput(RunWithArgs, app, "list")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output")
	os.RemoveAll("TestActionListAll")
}

func TestActionAppInfo(t *testing.T) {
	cfgPath := path.Join("TestActionAppInfo", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name:        "TestApp",
		Description: "test application",
		Host:        "http://123.456.789",
	})
	// tests
	expected := `TestApp:
	Description: test application
	Host: http://123.456.789
`
	app := app.New(cfgPath, http.Client{})
	out, err := captureOutput(RunWithArgs, app, "info", "app", "TestApp")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output")
	assert.Error(t, RunWithArgs(app, "info", "app"), "info app should generate an error with no args")
	os.RemoveAll("TestActionAppInfo")
}

func TestActionRequestInfo(t *testing.T) {
	cfgPath := path.Join("TestActionRequestInfo", ".sp9rk", "tests")
	// simulate apps being created
	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Method:      "GET",
		Description: "my request",
		Path:        "/path",
		Body:        "request body",
		Headers: []string{
			"X-API-Key: ABC123",
		},
	})
	// tests
	expected := `MyReq:
	description: my request
	method: GET
	path: /path
	headers:
	    - 'X-API-Key: ABC123'
	body: request body
`
	app := app.New(cfgPath, http.Client{})
	out, err := captureOutput(RunWithArgs, app, "info", "req", "-a", "TestApp", "MyReq")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output")
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	out, err = captureOutput(RunWithArgs, app, "info", "req", "MyReq")
	assert.NoError(t, err, "list req should NOT generate an error")
	assert.EqualValues(t, expected, out, "list should generate expected output with default app")
	assert.Error(t, RunWithArgs(app, "info", "req"), "info req should generate an error with no args")
	os.RemoveAll("TestActionRequestInfo")
}

type ClientMock struct{}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("Hello, world!")),
	}, nil
}
func TestActionCall(t *testing.T) {
	cfgPath := path.Join("TestActionCall", ".sp9rk", "tests")
	// simulate apps being created
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(t, "/path", r.URL.Path, "path is incorrect")
		assert.EqualValues(t, r.Header.Get("X-API-Key"), "ABC123", "headers are incorrect")
		assert.EqualValues(t, r.Method, "GET", "method is incorrect")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`Hello, World!`))
	}))
	defer server.Close()

	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
		Host: server.URL,
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Method:      "GET",
		Description: "my request",
		Path:        "/path",
		Body:        "request body",
		Headers: []string{
			"X-API-Key: ABC123",
		},
	})
	// test basic functionality
	app := app.New(cfgPath, http.Client{})
	assert.Error(t, RunWithArgs(app, "call"), "call should fail with no argument")
	assert.Error(t, RunWithArgs(app, "call", "MyReq"), "call should fail with no default app")
	assert.Error(t, RunWithArgs(app, "call", "../bad/path"), "call should fail invalid request name")
	assert.Error(t, RunWithArgs(app, "call", "-a", "../bad/app/path", "MyReq"), "call should fail with invalid app name")
	assert.Error(t, RunWithArgs(app, "call", "-a", "TestApp", "FakeReq"), "call should fail with unknown app")
	assert.Error(t, RunWithArgs(app, "call", "MyReq"), "call should fail with no default app")
	output, err := captureOutput(RunWithArgs, app, "call", "-a", "TestApp", "MyReq")
	assert.EqualValues(t, "Hello, World!\n", output, "call output should be response body")
	assert.NoError(t, err, "call should succeed with explicit app")
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	output, err = captureOutput(RunWithArgs, app, "call", "MyReq")
	assert.NoError(t, err, "call should succeed with default app")
	assert.EqualValues(t, "Hello, World!\n", output, "call output should be response body")

	// test extended functionality (flags)
	output, err = captureOutput(RunWithArgs, app, "call", "--verbose", "MyReq")
	assert.NoError(t, err, "call should succeed with verbose flag")
	respData := new(action.VerboseCallResponse)
	err = yaml.Unmarshal([]byte(output), respData)
	assert.NoError(t, err, "output should be valid yaml")
	assert.Equal(t, "GET "+server.URL+"/path", respData.Request, "request is incorrect")
	assert.NotEmpty(t, respData.Latency, "latency must not be empty")
	assert.Equal(t, "request body", respData.RequestBody, "request body is incorrect")
	assert.Equal(t, "200 OK", respData.Status, "response status is incorrect")
	assert.NotNil(t, respData.Headers, "request headers must not be nil")
	assert.Equal(t, "ABC123", respData.Headers["X-API-Key"], "request headers are incorrect")
	assert.Equal(t, "Hello, World!", respData.ResponseBody, "response body is incorrect")
	os.RemoveAll("TestActionCall")
}

func TestActionCallFail(t *testing.T) {
	cfgPath := path.Join("TestActionCallFail", ".sp9rk", "tests")
	// make a new server that will make every request return >=400 status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`Hello, World!`))
	}))
	defer server.Close()

	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
		Host: server.URL,
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Method:      "GET",
		Description: "my request",
		Path:        "/path",
		Body:        "request body",
		Headers: []string{
			"X-API-Key: ABC123",
		},
	})
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	app := app.New(cfgPath, http.Client{})
	output, err := captureOutput(RunWithArgs, app, "call", "--fail", "MyReq")
	assert.Error(t, err, "call should fail with fail flag")
	assert.Empty(t, output, "failed request should be silent with fail flag")
	os.RemoveAll("TestActionCallFail")
}

// TODO test redirects
func TestActionCallRedirects(t *testing.T) {
	cfgPath := path.Join("TestActionCallLocation", ".sp9rk", "tests")
	// make a new server that will make every request return >=400 status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hello" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`Hello, World!`))
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect-me" {
			w.Header().Add("Location", server.URL+"/hello")
			w.WriteHeader(http.StatusMovedPermanently)
			w.Write([]byte(`Don't follow me!`))
			return
		}
		w.WriteHeader(404)
	}))
	defer server2.Close()

	action.WriteAppFiles(cfgPath, &action.AppInfo{
		Name: "TestApp",
		Host: server2.URL,
	})
	action.WriteRequestFiles(cfgPath, "TestApp", &action.RequestInfo{
		Name:        "MyReq",
		Method:      "GET",
		Description: "my request",
		Path:        "/redirect-me",
	})
	os.WriteFile(path.Join(cfgPath, "current_app"), []byte("TestApp"), 0700)
	app := app.New(cfgPath, http.Client{})
	output, err := captureOutput(RunWithArgs, app, "call", "MyReq")
	assert.NoError(t, err, "call should not fail when following valid redirect")
	assert.Equal(t, "Hello, World!\n", output, "output should be equal to response body")
	output, err = captureOutput(RunWithArgs, app, "call", "--no-redirect", "MyReq")
	assert.NoError(t, err, "call should not fail when not following redirect")
	assert.Equal(t, "Don't follow me!\n", output, "output should be equal to response body")
	os.RemoveAll("TestActionCallLocation")
}

// helper functions for testing
func RunWithArgs(app *cli.App, args ...string) error {
	a := os.Args[0:1]
	a = append(a, args...)
	return app.Run(a)
}
func captureOutput(f func(*cli.App, ...string) error, app *cli.App, args ...string) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f(app, args...)
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}
