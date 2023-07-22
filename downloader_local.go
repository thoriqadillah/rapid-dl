package rapid

import "context"

// downloader that save the result into local file
type localDownloader struct {
	setting Setting
	ctx     context.Context
	cancel  context.CancelFunc
}

var Local = "local"

func newLocalDownloader(setting Setting) Downloader {
	ctx, cancel := context.WithCancel(context.Background())

	return &localDownloader{
		setting: setting,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (dl *localDownloader) Download(entry Entry) error {

	//TODO: implement download
	//TODO: handle filename duplication
	return nil
}

func (dl *localDownloader) Resume(entry Entry) error {
	//TODO: implement resume
	return nil
}

func (dl *localDownloader) Restart(entry Entry) error {
	//TODO: implement restart
	return nil
}

func (dl *localDownloader) Stop(entry Entry) error {
	dl.cancel()
	return nil
}

func init() {
	RegisterDownloader(Local, newLocalDownloader)
}
