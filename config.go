package ethutil

import (
	"log"
	"os/user"
	"path"
)

type LogType byte

const (
	LogTypeStdIn = 1
	LogTypeFile  = 2
)

// Config struct isn't exposed
type config struct {
	Db  Database
	Log *log.Logger

	ExecPath string
	Debug    bool
	Ver      string
}

var Config *config

// Read config doesn't read anything yet.
func ReadConfig() *config {
	if Config == nil {
		usr, _ := user.Current()
		path := path.Join(usr.HomeDir, ".ethereum")

		Config = &config{ExecPath: path, Debug: true, Ver: "0.0.1"}
	}

	return Config
}
