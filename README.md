# go-chat-app
A basic multi-client, multi room telnet chat server written in Go. 

## Build
```
go build app/buil.go
```

## Usage
* Starting server
```
go run app/main.go --config <toml_config_file_location>
```
* Connecting to the server:

```
telnet 127.0.0.1 8181
```

## Third party libraries

* github.com/sirupsen/logrus for logging
* github.com/BurntSushi/toml for parsing toml file
