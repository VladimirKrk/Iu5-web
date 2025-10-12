package config

import (
	"github.com/spf13/viper"
)

// Config хранит конфигурацию всего приложения
type Config struct {
	ServiceHost string
	ServicePort int
}

// NewConfig читает конфигурацию из файла config.toml
func NewConfig() (*Config, error) {
	// Указываем Viper, где искать и как называется наш конфиг
	viper.SetConfigName("config")   // имя файла без расширения
	viper.SetConfigType("toml")     // расширение
	viper.AddConfigPath("./config") // папка, где лежит файл

	// Читаем файл
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Создаем пустой объект конфига и "распаковываем" в него данные из файла
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
