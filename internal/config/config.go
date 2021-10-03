package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aerex/anki-cli/pkg/io"
	"github.com/spf13/viper"
)

const (
  XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
  ANKICLI_CONFIG_DIR = "ANKICLI_CONFIG_DIR"
)

const (
  AUTH_METHOD = "auth-method"
  EXEC = "exec"
)

type LoggerConfig struct {
  Level string `yaml:"logger.level"`
  File string `yaml:"logger.file"`
  Format string `yaml:"logger.format"`
}
type Config struct {
  // Options are `REST`
  Type string `yaml:"type"`
  // The URL to retrieve data from backend. If BackendType is `REST`
  Endpoint string `yaml:"endpoint"`
  // Options are `CUSTOM`, `PLAIN`, `STDIN`. 
  // CUSTOM will execute the command set in `pass-eval`.
  // PLAIN is a plain text password
  AuthMethod string `mapstructure:"auth-method"` 
  User string `yaml:"user"`
  PassEval string `mapstructure:"pass-eval"`
  Pass string  `yaml:"pass"`
  Logger LoggerConfig `yaml:"logger"`
  Dir string
}

func init() {
  viper.SetConfigName("config")
  viper.SetConfigType("yaml")
  viper.SetEnvPrefix("ankicli")
}

// Load the configuration file based on the this order of precedence
// 1. override configpath (ie. --config/-c flag)
// 2. env ANKI_CLI_CONFIG_DIR
// 3. env XDG_CONFIG_HOME
// 4. $XDG_CONFIG_HOME/anki-cli
// 5. AppData/anki-cli (Windows only)
// 6.  $HOME/.anki-cli
func Load(configPath string, config *Config) error {
  if configPath != "" {
    viper.AddConfigPath(configPath)
  } else if a := os.Getenv(ANKICLI_CONFIG_DIR); a != "" {
    viper.AddConfigPath(a)
  } else if x := os.Getenv(XDG_CONFIG_HOME); x != "" {
    viper.AddConfigPath(x)
  } else { 
    configDir, _ := GetUserConfig()
    viper.AddConfigPath(filepath.Join(configDir, "anki-cli"))
    homeDir, _ := os.UserHomeDir()
    viper.AddConfigPath(filepath.Join(homeDir, ".anki-cli"))
  }

  // Bind flags and environment variables to override configs
  //viper.BindPFlags(cmd.Flags())
  viper.AutomaticEnv();

  err := viper.ReadInConfig();
  if err != nil {
    return err
  }

  erru := viper.Unmarshal(&config)
  if erru != nil {
    return erru
  }

  if config.AuthMethod == EXEC {
    out, err := io.Eval(config.Pass)
    if err != nil {
      return err
    }
    config.Pass = out
  }

  // Remove trailing / from endpoint if there is one
  if config.Endpoint != "" {
    lastCharIdx := strings.LastIndex( config.Endpoint, "/")
    if config.Endpoint[lastCharIdx:] == "/"  && len(config.Endpoint) == lastCharIdx + 1 {
      config.Endpoint = config.Endpoint[0:lastCharIdx]
    }
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
func GenerateSampleConfig(config *Config) (string, error) {
  // Create config directory if it doesn't exist
  userConfig, _ := GetUserConfig()
  configDir := filepath.Join(userConfig, "anki-cli")
  if _, err := os.Stat(configDir); os.IsNotExist(err) {
    os.Mkdir(configDir, 0700)
  }

  // TODO: export to helper method if used again
  _, fileName, _, _ := runtime.Caller(0)
  moduleDir := filepath.Join(filepath.Dir(fileName))

  // Copy over sample config to configuration directory
  sampleConfigFile, err := ioutil.ReadFile(filepath.Join(moduleDir, "../../configs/config"))
  if err != nil {
    return "", err
  }
  configFilePath := filepath.Join(configDir, "config")
  err = ioutil.WriteFile(configFilePath, sampleConfigFile,  0700)
  if err != nil {
    return "", err
  }

  // Load new configuration 
  err = Load(configDir, config)
  if err != nil {
    return "", err
  }

  return configFilePath, nil
}

