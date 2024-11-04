package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aerex/go-anki/pkg/io"
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	XDG_CONFIG_HOME    = "XDG_CONFIG_HOME"
	ANKICLI_CONFIG_DIR = "ANKICLI_CONFIG_DIR"
)

const (
	AUTH_METHOD = "auth-method"
	EXEC        = "exec"
)

type Logger struct {
  Level  string `toml:"level" comment:"The log level verbosity. Default is DEBUG. Available levels: TRACE, DEBUG, INFO, WARN, ERROR, FATAL, OFF"`
  File   string `toml:"file" comment:"The file path where the logs will be stored"`
	Sql    bool   `toml:"sql" comment:"Enable to log SQL statements"`
}

type Prompt struct {
	Vim bool `toml:"vim" comment:"Enable vim mode to use j/k to cycle through selections"`
}

type Color struct {
  Hint string `toml:"hint" comment:"The color to use for hints in cloze"`
}

type API struct {
	User     string `toml:"user"`
  PassEval string `toml:"pass-cmd,inline" comment:"Run a command to retrieve the password"`
	Pass     string `toml:"pass,omitempty"`
	// The URL to retrieve data from backend.
	Endpoint string `toml:"endpoint" comment:"URL to retrieve data from backend"`
}

type General struct {
	// Options are `REST` and `DB`
  Type string `toml:"type" mapstructure:"type" comment:"Options are REST and DB"`
	// SchedulerVersion sets the the scheduler version to use when syncing. Options are 2 or 3
	// @see https://faqs.ankiweb.net/the-anki-2.1-scheduler.html and https://faqs.ankiweb.net/the-2021-scheduler.html
	// for information on compatibility
	SchedulerVersion int `toml:"sched" mapstructure:"sched" comment:"Sets the scheduler version to use when syncing. Options are 2 or 3"`
	// The path of for the editor that will be launched when editing content (ie: vim or notepad)
	// Default will use the editor set by the EDITOR or ANKICLI_EDITOR environment variable
	Editor string `toml:"editor" comment:"The path of for the editor that will be launched when editing content (ie: vim or notepad)"`
}

type DB struct {
	// the database driver (ie: sqlite3)
  Driver string `toml:"driver" comment:"the database driver: Options are sqlite3"`
	// location of database file
	File string `toml:"file" comment:"location of database file"`
}
type Config struct {
	// (Optional) the location of the database. If set TYPE must be set as DB
	DB DB `toml:"db,omitempty" comment:"the location of the database. If set TYPE must be set as DB"`
	// The username credential to access backend
	Logger  Logger  `toml:"logger"`
	API     API     `toml:"api"`
	General General `toml:"general"`
	Color   Color   `toml:"color,omitempty"`
	Dir     string  `toml:"dir,omitempty"`
	Prompt  Prompt  `toml:"prompt"`
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.SetEnvPrefix("ankicli")
}

// TODO: Is there a better way of doing this?
// Maybe just define the config as a struct
// instead of loading it
// Load the sample config test
func LoadSampleConfig() (Config, error) {
	var cfg Config
	_, fileName, _, _ := runtime.Caller(0)
	moduleDir := filepath.Join(filepath.Dir(fileName))
	path := filepath.Join(moduleDir, "../../configs")
	if err := Load(path, &cfg, &io.IO{ExecContext: exec.Command}); err != nil {
		fmt.Println("Failed to load config")
		return cfg, err
	}
	return cfg, nil
}

// Load the configuration file based on the this order of precedence
// 1. override configpath (ie. --config/-c flag)
// 2. env ANKI_CLI_CONFIG_DIR
// 3. env XDG_CONFIG_HOME
// 4. $XDG_CONFIG_HOME/anki-cli
// 5. AppData/anki-cli (Windows only)
// 6. $HOME/.anki-cli
func Load(configPath string, config *Config, io *io.IO) error {
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else if a := os.Getenv(ANKICLI_CONFIG_DIR); a != "" {
		viper.AddConfigPath(a)
	} else if x := os.Getenv(XDG_CONFIG_HOME); x != "" {
		viper.AddConfigPath(filepath.Join(x, "anki-cli"))
	} else {
		configDir, _ := GetUserConfig()
		viper.AddConfigPath(filepath.Join(configDir, "anki-cli"))
		homeDir, _ := homedir.Dir()
		viper.AddConfigPath(filepath.Join(homeDir, ".anki-cli"))
	}

	// Bind flags and environment variables to override configs
	//viper.BindPFlags(cmd.Flags())
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	erru := viper.Unmarshal(&config)
	if erru != nil {
		return erru
	}

	if config.API.PassEval != "" {
		buf := bytes.NewBufferString("")
		err := io.Eval(config.API.PassEval, buf)
		if err != nil {
			return err
		}
		config.API.Pass = strings.Replace(buf.String(), "\n", "", 1)
	}

	// Remove trailing / from endpoint if there is one
	if config.API.Endpoint != "" {
		lastCharIdx := strings.LastIndex(config.API.Endpoint, "/")
		if config.API.Endpoint[lastCharIdx:] == "/" && len(config.API.Endpoint) == lastCharIdx+1 {
			config.API.Endpoint = config.API.Endpoint[0:lastCharIdx]
		}
	}

	if config.DB.File != "" {
		expandedPath, err := homedir.Expand(config.DB.File)
		if err != nil {
			return err
		}
		config.DB.File = expandedPath
	}

	// Retrieve config file path from absolute file path
	configFilePathUsed := viper.ConfigFileUsed()
	lastSlashIdx := strings.LastIndex(configFilePathUsed, "/")
	config.Dir = configFilePathUsed[0:(lastSlashIdx)]

	return nil
}

// Return usr configuration directory.
// If Unix or MacOs return $HOME/.config
// otherwise return directory using os.UserConfigDir()
func GetUserConfig() (string, error) {
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir + "/.config", nil
	}
	return os.UserConfigDir()
}

// Copy over sample configuration to user configuration directory
// If user configuration directory does not exist, create it
// If MacOs/Unix, configuration file will be saved under $HOME/.config/anki-cli/config
// If Windows, %USERPROFILE/anki-cli/config
func GenerateSampleConfig(config *Config, io *io.IO) (string, error) {
	// Create config directory if it doesn't exist
	userConfig, _ := GetUserConfig()
	configDir := filepath.Join(userConfig, "anki-cli")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0700); err != nil {
			return "", err
		}
	}

	config.General.SchedulerVersion = 2
  config.Logger.Level = strings.ToUpper(zerolog.DebugLevel.String())

	out, err := toml.Marshal(config)
	if err != nil {
		return "", err
	}

	configFilePath := filepath.Join(configDir, "config.toml")
	err = os.WriteFile(configFilePath, out, 0600)
	if err != nil {
		return "", err
	}

	// Load new configuration
	err = Load(configDir, config, io)
	if err != nil {
		return "", err
	}

	return configFilePath, nil
}
