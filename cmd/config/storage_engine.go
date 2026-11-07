package config

import (
	"fmt"
	"net/http"

	"github.com/fitant/storage-engine-go/storageengine"
	"github.com/spf13/viper"
)

type storageEngineConfig struct {
	URL            string
	ObjectID       string
	ObjectPassword string
}

func loadSEConfig() (*storageEngineConfig, error) {
	v := viper.New()

	v.BindEnv("url", "SE_URL")
	v.BindEnv("object_id", "SE_OBJ_ID")
	v.BindEnv("object_password", "SE_OBJ_PASS")

	se := storageEngineConfig{
		URL:            v.GetString("url"),
		ObjectID:       v.GetString("object_id"),
		ObjectPassword: v.GetString("object_password"),
	}
	if se.URL == "" || se.ObjectID == "" || se.ObjectPassword == "" {
		return nil, fmt.Errorf("Storage Engine config was not supplied")
	}
	return &se, nil
}

func initSEObj() (*storageengine.Object, error) {
	seCfg, err := loadSEConfig()
	if err != nil {
		return nil, err
	}

	seClient, err := storageengine.NewClientConfig(http.DefaultClient, seCfg.URL)
	if err != nil {
		return nil, err
	}
	configObject, err := storageengine.NewObject(seClient)
	if err != nil {
		return nil, err
	}

	configObject.SetID(seCfg.ObjectID)
	configObject.SetPassword(seCfg.ObjectPassword)
	if err := configObject.Refresh(); err != nil {
		return nil, err
	}
	if configObject.GetData() == "" {
		return nil, fmt.Errorf("SE Object has no data")
	}
	return configObject, nil
}
