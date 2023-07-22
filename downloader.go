package rapid

type (

	// Downloader is interface to perform a download, pause, resume, restart, and stop for certain download
	Downloader interface {
		Download(entry Entry) error
		Resume(entry Entry) error
		Restart(entry Entry) error
		Stop(entry Entry) error
	}

	// DownloaderFunc is an abstract for creating a Downloader
	DownloaderFunc func(s Setting) Downloader
)

var downloadermap = make(map[string]DownloaderFunc)

func NewDownloader(provider string, setting Setting) Downloader {
	downloader := downloadermap[provider]
	return downloader(setting)
}

func RegisterDownloader(name string, impl DownloaderFunc) {
	downloadermap[name] = impl
}
