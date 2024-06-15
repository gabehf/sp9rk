This project is under active development and may have breaking API changes! Use at your own risk.
# Sp9rk - Easily save and call web requests
Pronounced "spark".

Sp9rk allows you to easily register and execute complex web requests without having to rewrite long cUrl commands. Just create the request once and execute as many times as you need.
## Install
### From Source
Ensure you have go version 1.22.2 or greater.

Clone the repo
```bash
git clone https://github.com/gabehf/sp9rk
cd sp9rk
```
Run `go build`
```bash
go build
```
Do whatever you want with the executable
```bash
$ sp9rk -v
sp9rk version v0.0.1
```
## Usage
### Create
Register an application
```bash
$ sp9rk create app -u http://localhost:8080 ExampleApp
Created application ExampleApp
```
Register a request for the app
```bash
$ sp9rk create req \
  -a ExampleApp \
  -p "/path" \
  -b "{'body':'request body'}" \
  -X POST \
  MyRequest
Created request MyRequest
```
## Switch
You can set the default application your commands effect using `switch`
```bash
$ sp9rk switch ExampleApp
ExampleApp
```
Now sp9rk will default to using that application, instead of needing to specify the app with the `-a` flag
## Call
Call the request
```bash
$ sp9rk call MyRequest
Hello, World!
```
You can also return more information with the `--verbose -v` flag when making a call
```bash
$ sp9rk call -v MyRequest
Request: GET http://localhost:8080/path
RequestBody: "{'body':'request body'}"
Headers: {}
Latency: 103.731894ms
Status: 200 OK
ResponseBody: Hello, World!
```
## Edit
You can edit the definitions of existing requests or apps
```bash
$ sp9rk edit req --method POST --path /other-path MyRequest

$ sp9rk call -v MyRequest
Request: POST http://localhost:8080/other-path
RequestBody: "{'body':'request body'}"
Headers: {}
Latency: 67.82371ms
Status: 200 OK
ResponseBody: Hello again, World!
```
## Delete
You can delete apps or requests using `delete app/req`
```bash
$ sp9rk delete req MyRequest
You are about to delete the request MyRequest in application ExampleApp.
This action cannot be undone.
Are you sure? [y/N]: y
request MyRequest has been deleted 
```

# TODO
- [ ] Allow users to specify the number of redirects to follow before stopping
- [ ] Allow flags to be saved along with requests
- [ ] Allow parameters to be used inside both requests paths and bodies
- [ ] Allow for requests to use files as request bodies
- [ ] Allow for commands flags to be used both before and after arguments (i.e. allowing `sp9rk create req MyReq -a MyApp` as well as `sp9rk create -a MyApp MyReq`)
- [ ] Add easy install script and/or package

# Albums that fueled development
| Album                       | Artist           |
|-----------------------------|------------------|
| Emotion                     | Carly Rae Jepsen |
| To Pimp a Butterfly         | Kendrick Lamar   |
| good kid, m.A.A.d city      | Kendrick Lamar   |
| if i could make it go quiet | girl in red      |
| DROP                        | Minami (美波)     |