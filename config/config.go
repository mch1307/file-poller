package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Conf holds app conf
var Conf conf

// Conf holds the server config
type conf struct {
	SourceDir string `toml:"SourceDir"`
	DestDir   string `toml:"DestDir"`
	LogDir    string `toml:"LogDir"`
	LogLevel  string `toml:"LogLevel"`
}

// coalesce returns the first non nil/zero passed value as string
// numeric 0 value considered as empty
func coalesce(data ...interface{}) string {
	for _, v := range data {
		switch v := v.(type) {
		case string:
			if len(v) > 0 {
				return v
			}
		case int:
			if v > 0 {
				strconv.Itoa(v)
				return strconv.Itoa(v)
			}
		}
	}
	return ""
}

// Initialize populates the Conf variable
func Initialize(cfg string) {
	// load config file if any
	if cfg != "" {
		if _, err := os.Stat(cfg); err != nil {
			wrkDir, _ := os.Getwd()
			Conf.LogDir = wrkDir
		} else {
			if _, err := toml.DecodeFile(cfg, &Conf); err != nil {
				fmt.Println("Error parsing config file: ", err)
			}
		}
	}

	fmt.Printf("Starting with config: %+v", Conf)
	fmt.Println(" ")
}
