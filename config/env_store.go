package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/ovatu/redalert/checks"
	"github.com/ovatu/redalert/notifiers"
)

type EnvStore struct {
	name	 string
	data     EnvStoreData
}

type EnvStoreData struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
	Preferences   Preferences        `json:"preferences"`
}

func NewEnvStore(name string) (*EnvStore, error) {
	config := &EnvStore{name: name}
	err := config.read()
	if err != nil {
		return nil, err
	}

	// create check ID if not present
	for i := range config.data.Checks {
		if config.data.Checks[i].ID == "" {
			config.data.Checks[i].ID = generateID(8)
		}
	}

	// create notification ID if not present
	for i := range config.data.Notifications {
		if config.data.Notifications[i].ID == "" {
			config.data.Notifications[i].ID = generateID(8)
		}
	}

	return config, nil
}

func (f *EnvStore) read() error {
	file := os.Getenv(f.name)

	if file == "" {
		return errors.New("environment variable empty")
	}

	var data EnvStoreData
	err := json.Unmarshal([]byte(file), &data)
	if err != nil {
		return err
	}
	f.data = data
	return nil
}

func (f *EnvStore) Notifications() ([]notifiers.Config, error) {
	return f.data.Notifications, nil
}

func (f *EnvStore) Checks() ([]checks.Config, error) {
	return f.data.Checks, nil
}

func (f *EnvStore) Preferences() (Preferences, error) {
	return f.data.Preferences, nil
}
