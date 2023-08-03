package rapid

import (
	"os"
	"testing"
)

const dummypdf = "https://file-examples.com/storage/fee3d1095964bab199aee29/2017/04/file_example_MP4_480_1_5MG.mp4"

func TestDownloadLocalSuccess(t *testing.T) {
	entry, err := Fetch(dummypdf, DefaultSetting())
	if err != nil {
		t.Error("Error while fetching dummy video:", err.Error())
	}

	downloader := NewDownloader(DownloaderDefault, DefaultSetting())
	if err := downloader.Download(entry); err != nil {
		t.Error("Error while downloading dummy video:", err.Error())
	}

	file, err := os.Stat(entry.Location())
	if err != nil {
		t.Error("Errow downloading file:", err.Error())
	}

	if file.Size() != entry.Size() {
		t.Errorf("Download has different size. Expected %d, but got %d", entry.Size(), file.Size())
	}
}
