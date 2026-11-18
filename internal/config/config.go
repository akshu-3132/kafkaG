package config

// Config for the system
type Config struct {
	StorageConfig StroageConfig
}

type StroageConfig struct {
	LogFilePath   string
	IndexFilePath string
	IndexInterval int64
}

func DefaultConfig() *Config {
	return &Config{
		StorageConfig: StroageConfig{
			LogFilePath:   "data/logs/logfile.log",
			IndexFilePath: "data/logs/indexfile.idx",
			IndexInterval: 1000, // in bytes
		},
	}
}
