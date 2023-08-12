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

	// DownloaderFactory is an abstract for creating a Downloader
	DownloaderFactory func(o *downloaderOption) Downloader

	OnProgress func(...interface{})

	downloaderOption struct {
		setting Setting
	}

	DownloaderOptions func(o *downloaderOption)
)

func SetDownloaderSetting(setting Setting) DownloaderOptions {
	return func(o *downloaderOption) {
		o.setting = setting
	}
}

var downloadermap = make(map[string]DownloaderFactory)

func NewDownloader(provider string, options ...DownloaderOptions) Downloader {
	opt := &downloaderOption{
		setting: DefaultSetting(),
	}

	for _, option := range options {
		option(opt)
	}

	downloader, ok := downloadermap[provider]
	if !ok {
		log.Panicf("Provider %s is not implemented", provider)
		return nil
	}

	return downloader(opt)
}

func RegisterDownloader(name string, impl DownloaderFactory) {
	downloadermap[name] = impl
}
