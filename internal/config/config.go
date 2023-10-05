package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/mitchellh/go-homedir"
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

type LoggerConfig struct {
	Level  string `yaml:"logger.level"`
	File   string `yaml:"logger.file"`
	Format string `yaml:"logger.format"`
}
type ColorConfig struct {
	Hint string `yaml:"color.hint"`
}

type DBConfig struct {
	// the database driver (ie: sqlite3)
	Driver string `yaml:"db.driver" toml:"general.type"`
	// location of db
	Path string `yaml:"db.path"`
}
type Config struct {
	// Options are `REST`, `DB`
	Type string `yaml:"type"`
	// (Optional) the location of the database. If set TYPE must be set as DB
	DB DBConfig `yaml:"db,omitempty"`
	// SchedulerVersion sets the the scheduler version to use when syncing. Options are 2 or 3
	// @see https://faqs.ankiweb.net/the-anki-2.1-scheduler.html and https://faqs.ankiweb.net/the-2021-scheduler.html
	// for informtation on compatibility
	SchedulerVersion int `yaml:"sched" toml:"general.sched"`
	// The path of for the editor that will be launched when editing content (ie: vim or notepad)
	// Default will use the editor set by the EDITOR or ANKICLI_EDITOR environment variable
	Editor string `yaml:"editor"`
	// The URL to retrieve data from backend. If BackendType is `REST`
	Endpoint string `yaml:"endpoint"`
	// The username credential to access backend
	User     string       `yaml:"user"`
	PassEval string       `mapstructure:"passEval"`
	Pass     string       `yaml:"pass"`
	Logger   LoggerConfig `yaml:"logger"`
	Dir      string
	Color    ColorConfig `yaml:"color,omitempty"`
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
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
// 6.  $HOME/.anki-cli
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

	if config.PassEval != "" {
		buf := bytes.NewBufferString("")
		err := io.Eval(config.PassEval, buf)
		if err != nil {
			return err
		}
		config.Pass = strings.Replace(buf.String(), "\n", "", 1)
	}

	// Remove trailing / from endpoint if there is one
	if config.Endpoint != "" {
		lastCharIdx := strings.LastIndex(config.Endpoint, "/")
		if config.Endpoint[lastCharIdx:] == "/" && len(config.Endpoint) == lastCharIdx+1 {
			config.Endpoint = config.Endpoint[0:lastCharIdx]
		}
	}

	if config.DB.Path != "" {
		expandedPath, err := homedir.Expand(config.DB.Path)
		if err != nil {
			return err
		}
		config.DB.Path = expandedPath
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
		os.Mkdir(configDir, 0700)
	}

	moduleDir := utils.CurrentModuleDir()

	// Copy over sample config to configuration directory
	sampleConfigFile, err := ioutil.ReadFile(filepath.Join(moduleDir, "../../configs/config"))
	if err != nil {
		return "", err
	}
	configFilePath := filepath.Join(configDir, "config")
	err = ioutil.WriteFile(configFilePath, sampleConfigFile, 0700)
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
