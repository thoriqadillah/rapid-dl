package rapid

import (
	"os"
	"path/filepath"
	"testing"
)

const dummypdf = "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"

func TestDownloadLocalSuccess(t *testing.T) {
	entry, err := Fetch(dummypdf, DefaultSetting())
	if err != nil {
		t.Error("Error while fetching dummy pdf:", err.Error())
	}

	downloader := NewDownloader(DownloaderLocal, DefaultSetting())
	if err := downloader.Download(entry); err != nil {
		t.Error("Error while downloading dummy pdf:", err.Error())
	}

	if err := os.Remove(entry.Location()); err != nil {
		t.Error("Error removing dummy pdf:", err.Error())
	}
}

func TestHandleDuplicate(t *testing.T) {
	entry, err := Fetch(dummypdf, DefaultSetting())
	if err != nil {
		t.Error("Error while fetching dummy pdf:", err.Error())
	}

	downloader := NewDownloader(DownloaderLocal, DefaultSetting())
	if err := downloader.Download(entry); err != nil {
		t.Error("Error while downloading dummy pdf:", err.Error())
	}

	if err := downloader.Download(entry); err != nil {
		t.Error("Error while downloading dummy pdf:", err.Error())
	}

	home, _ := os.UserHomeDir()
	dummy1 := filepath.Join(home, "Downloads", "dummy.pdf")
	dummy2 := filepath.Join(home, "Downloads", "dummy (1).pdf")

	for _, r := range []string{dummy1, dummy2} {
		if _, err := os.Stat(r); err != nil {
			t.Error("Download failed")
		}

		if err := os.Remove(r); err != nil {
			t.Error("Error removing dummy pdf:", err.Error())
		}
	}

}
