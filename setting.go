package rapid

import (
	"os"
	"path/filepath"
)

type (
	Setting interface {
		// location where the download will be placed
		DownloadLocation() string

		// location where the data for this application will be stored
		DataLocation() string

		// max number of file will be downloaded at the same time
		MaxConcurrentDownload() int

		// max retry will be executed when there is an error while downloading
		MaxRetry() int

		// logger provider that will be used to log something, e.g file, std, etc
		LoggerProvider() string

		// minimum size in MB for a chunk
		MinChunkSize() int64
	}

	setting struct {
		downloadLocation      string
		dataLocation          string
		maxConcurrentDownload int
		maxRetry              int
		loggerProvider        string
		poolsize              int
		minChunkSize          int64
	}
)

func DefaultSetting() Setting {
	home, _ := os.UserHomeDir()

	// location
	data := filepath.Join(home, ".rapid")
	download := filepath.Join(home, "Downloads")

	os.MkdirAll(data, os.ModePerm)

	return &setting{
		downloadLocation:      download,
		dataLocation:          data,
		maxConcurrentDownload: 4,
		maxRetry:              3,
		loggerProvider:        LoggerStdOut,
		minChunkSize:          1024 * 1024 * 5, // 5 MB
	}
}

func (s *setting) DownloadLocation() string {
	return s.downloadLocation
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

func (s *setting) MinChunkSize() int64 {
	return s.minChunkSize
}
