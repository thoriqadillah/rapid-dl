package rapid

import "log"

type (
	// Downloader is interface to perform a download, pause, resume, restart, and stop for certain download
	Downloader interface {
		Download(entry Entry) error
		Resume(entry Entry) error
		Restart(entry Entry) error
		Stop(entry Entry) error
	}

	Watcher interface {
		Watch(update OnProgress)
	}

	// DownloaderFunc is an abstract for creating a Downloader
	DownloaderFunc func(s Setting) Downloader

	OnProgress func(...interface{})
)

var downloadermap = make(map[string]DownloaderFunc)

func NewDownloader(provider string, setting Setting) Downloader {
	downloader, ok := downloadermap[provider]
	if !ok {
		log.Panicf("Provider %s is not implemented", provider)
		return nil
	}

	return downloader(setting)
}

func RegisterDownloader(name string, impl DownloaderFunc) {
	downloadermap[name] = impl
}
