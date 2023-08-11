package rapid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadLocalOneChunkSuccess(t *testing.T) {
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

	os.Remove(entry.Location())
}

func TestDownloadLocalMultipleChunkSuccess(t *testing.T) {
	url := "https://link.testfile.org/PDF50MB"
	entry, err := Fetch(url, DefaultSetting())
	if err != nil {
		t.Error("Error fetching dummy video:", err.Error())
	}

	downloader := NewDownloader(DownloaderDefault, DefaultSetting())
	if err := downloader.Download(entry); err != nil {
		t.Error("Error downloading dummy video:", err.Error())
	}

	if watcher, ok := downloader.(Watcher); ok {
		watcher.Watch(func(i ...interface{}) {
			log.Println(i)
		})
	}

	file, err := os.Stat(entry.Location())
	if err != nil {
		t.Error("Errow downloading file:", err.Error())
	}

	if file.Size() != entry.Size() {
		t.Errorf("Download has different size. Expected %d, but got %d", entry.Size(), file.Size())
	}

	os.Remove(entry.Location())
}

func TestStopDownloadLocalOneChunkSuccess(t *testing.T) {
	url := "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=SampleDocs-Test%20PDF%20File%20With%20Dummy%20Data%20For%20Testing&ext=pdf"

	setting := DefaultSetting()
	entry, err := Fetch(url, setting)
	if err != nil {
		t.Error("Error fetching dummy video:", err.Error())
	}

	downloader := NewDownloader(DownloaderDefault, setting)
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

	for i := 0; i < entry.ChunkLen(); i++ {
		chunkfile := filepath.Join(setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), i))
		if err := os.Remove(chunkfile); err != nil {
			t.Error("Error removing chunk file")
		}
	}
}

func TestStopDownloadLocalMultipleChunkSuccess(t *testing.T) {
	url := "https://link.testfile.org/PDF50MB"
	setting := DefaultSetting()
	entry, err := Fetch(url, setting)
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

	time.Sleep(2 * time.Second)

	if err := downloader.Stop(entry); err != nil {
		t.Error("Error stopping download:", err.Error())
	}

	for i := 0; i < entry.ChunkLen(); i++ {
		chunkfile := filepath.Join(setting.DownloadLocation(), fmt.Sprintf("%s-%d", entry.ID(), i))
		if err := os.Remove(chunkfile); err != nil {
			t.Error("Error removing chunk file")
		}
	}
}

func TestStopDownloadLocalResumeSuccess(t *testing.T) {
	url := "https://link.testfile.org/PDF50MB"
	setting := DefaultSetting()
	entry, err := Fetch(url, setting)
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

	time.Sleep(2 * time.Second)

	if err := downloader.Stop(entry); err != nil {
		t.Error("Error stopping download:", err.Error())
	}

	if err := downloader.Resume(entry); err != nil {
		t.Error("Error resuming download:", err.Error())
	}

	file, err := os.Stat(entry.Location())
	if err != nil {
		t.Error(err)
	}

	if file.Size() != entry.Size() {
		t.Errorf("Downloaded file size and entry file size is different. Expected to be %d, but got %d. Minus %d MB", entry.Size(), file.Size(), (entry.Size()-file.Size())/(1024*1024))
	}

	os.Remove(entry.Location())
}
