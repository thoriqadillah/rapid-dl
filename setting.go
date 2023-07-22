package rapid

import (
	"os"
	"path/filepath"
)

type Setting interface {
	DownloadLocation() string
	DataLocation() string
	MaxConcurrentDownload() int
	MaxRetry() int
	LoggerProvider() string
}

type setting struct {
	downloadLocation      string
	dataLocation          string
	maxConcurrentDownload int
	maxRetry              int
	loggerProvider        string
}

func DefaultSetting() Setting {
	home, _ := os.UserHomeDir()

	// location
	data := filepath.Join(home, ".gown")
	download := filepath.Join(home, "Downloads")

	return &setting{
		downloadLocation:      download,
		dataLocation:          data,
		maxConcurrentDownload: 4,
		maxRetry:              3,
		loggerProvider:        stdLog,
	}
}

func (s *setting) DownloadLocation() string {
	return s.dataLocation
}

func (s *setting) DataLocation() string {
	return s.dataLocation
}

func (s *setting) MaxConcurrentDownload() int {
	return s.maxConcurrentDownload
}

func (s *setting) MaxRetry() int {
	return s.maxRetry
}

func (s *setting) LoggerProvider() string {
	return s.loggerProvider
}