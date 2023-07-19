package rapid

type (
	Downloader interface {
		Download() error
		Pause() error
		Resume() error
		Restart() error
	}

	downloader struct {
		url string
	}
)

func New(url string) Downloader {
	return &downloader{
		url: url,
	}
}

func (d *downloader) Download() error
func (d *downloader) Pause() error
func (d *downloader) Resume() error
func (d *downloader) Restart() error
