package rapid

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestDownloadLocalSuccess(t *testing.T) {
	url := "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=SampleDocs-Test%20PDF%20File%20With%20Dummy%20Data%20For%20Testing&ext=pdf"
	entry, err := Fetch(url, DefaultSetting())
	if err != nil {
		t.Error("Error fetching dummy video:", err.Error())
	}

	downloader := NewDownloader(DownloaderDefault, DefaultSetting())
	if err := downloader.Download(entry); err != nil {
		t.Error("Error downloading dummy video:", err.Error())
	}

	file, err := os.Stat(entry.Location())
	if err != nil {
		t.Error("Errow downloading file:", err.Error())
	}

	if file.Size() != entry.Size() {
		t.Errorf("Download has different size. Expected %d, but got %d", entry.Size(), file.Size())
	}
}

func TestStopDownloadLocalSuccess(t *testing.T) {
	url := "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=SampleDocs-Test%20PDF%20File%20With%20Dummy%20Data%20For%20Testing&ext=pdf"
	entry, err := Fetch(url, DefaultSetting())
	if err != nil {
		t.Error("Error fetching dummy video:", err.Error())
	}

	downloader := NewDownloader(DownloaderDefault, DefaultSetting())
	if watcher, ok := downloader.(Watcher); ok {
		watcher.Watch(func(i ...interface{}) {
			log.Println(i)
		})
	}

	go func() {
		if err := downloader.Download(entry); err != nil {
			t.Error("Error downloading dummy video:", err.Error())
		}
	}()

	time.Sleep(5 * time.Second)

	if err := downloader.Stop(entry); err != nil {
		t.Error("Error stopping download:", err.Error())
	}
}
