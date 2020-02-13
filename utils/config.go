package utils

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// CronoscopeConfig is the application configuration
type CronoscopeConfig struct {
	Disabled            bool   `default: false`
	PollingInterval     int    `default:"10" split_words:"true"`
	PushergatewayHost   string `required:"true" split_words:"true"`
	PushergatewayPort   int    `default:"9091" split_words:"true"`
	PushRetries         int    `default:"3" split_words:"true"`
	PushRetriesInterval int    `default:"2" split_words:"true"`
	LabelJob            string `required:"true" split_words:"true"`
	Labels              map[string]string
}

const CRONOSCOPE_LABELS_PREFIX = "CRONOSCOPE_LABEL_"

func readLabels(config *CronoscopeConfig) {
	config.Labels = make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		label := strings.ToUpper(pair[0])
		value := pair[1]

		if strings.HasPrefix(pair[0], CRONOSCOPE_LABELS_PREFIX) {
			label = strings.ToLower(strings.TrimPrefix(label, CRONOSCOPE_LABELS_PREFIX))
			config.Labels[label] = value
		}

	}

	// Because these are already top level methdos
	delete(config.Labels, "job")
}

// ReadConfig reads the configuration from enviornment and validates.
// In case of an error, this function will quit the program
func ReadConfig() CronoscopeConfig {

	var config CronoscopeConfig
	readLabels(&config)
	return config

	// 1. If the program to be executed is missing, no point in
	// looking at environment variables.
	if len(os.Args) < 2 {
		fmt.Println("usage: cronosocpe your-command-here")
		os.Exit(1)
	}

	isDisabled, isSet := os.LookupEnv("CRONOSCOPE_DISABLED")

	// If disabled, no point in processing environment variables
	if isSet == true && strings.ToLower(isDisabled) == "true" {
		config.Disabled = true
		log.Printf("CRONOSCOPE_DISABLED was set to true. Not processing further environment variables.")
		return config
	}

	// If required environment variables are not present,
	// no point in moving forward
	err := envconfig.Process("CRONOSCOPE", &config)

	if err != nil {
		log.Fatalf(err.Error())
	}

	retryTime := (config.PushRetriesInterval * config.PushRetries) + 3

	if config.PollingInterval < retryTime {
		fmt.Println("CRONOSCOPE_POLLING_INTERVAL is too short")
		fmt.Println("CRONOSCOPE_POLLING_INTERVAL should be at least 2 seconds more than product of CRONOSCOPE_PUSH_RETRIES x CRONOSCOPE_PUSH_RETRIES_INTERVAL")

		fmt.Println("CRONOSCOPE_POLLING_INTERVAL=", config.PollingInterval)
		fmt.Println("CRONOSCOPE_PUSH_RETRIES=", config.PushRetries)
		fmt.Println("CRONOSCOPE_PUSH_RETRIES_INTERVAL=", config.PushRetriesInterval)
		os.Exit(1)
	}
	return config

}
