package ethutil

import (
	"log"
	"os/user"
	"path"
)

// Config struct isn't exposed
type config struct {
	Db  Database
	Log *log.Logger

	ExecPath string
}

var Config *config

// Read config doesn't read anything yet.
func ReadConfig() *config {
	if Config == nil {
		usr, _ := user.Current()
		path := path.Join(usr.HomeDir, ".ethereum")

		Config = &config{ExecPath: path}
	}

	return Config
}
