package rapid

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestFilename(t *testing.T) {
	link := "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
	req, err := http.Get(link)
	if err != nil {
		t.Error("Error fetching link:", err.Error())
	}

	name := filename(req)
	if name != "dummy.pdf" {
		t.Errorf("Error fetching file name. Expected dummy.pdf, but got %s", name)
	}

	link = "https://freetestdata.com/wp-content/uploads/2021/09/Free_Test_Data_100KB_PDF.pdf"
	req, err = http.Get(link)
	if err != nil {
		t.Error("Error fetching link:", err.Error())
	}

	name = filename(req)
	if name != "Free_Test_Data_100KB_PDF.pdf" {
		t.Errorf("Error fetching file name. Expected dummy.pdf, but got %s", name)
	}

	link = "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=SampleDocs-sample-pdf-file&ext=pdf"
	req, err = http.Get(link)
	if err != nil {
		t.Error("Error fetching link:", err.Error())
	}

	name = filename(req)
	if name != "SampleDocs-sample-pdf-file.pdf" {
		t.Errorf("Error fetching file name. Expected dummy.pdf, but got %s", name)
	}

	link = "https://research.nhm.org/pdfs/10592/10592-002.pdf"
	req, err = http.Get(link)
	if err != nil {
		t.Error("Error fetching link:", err.Error())
	}

	name = filename(req)
	if name != "10592-002.pdf" {
		t.Errorf("Error fetching file name. Expected dummy.pdf, but got %s", name)
	}
}

func TestHandleDuplicateName(t *testing.T) {
	home, _ := os.UserHomeDir()
	name := filepath.Join(home, "Downloads", "test.pdf")

	newname := handleDuplicate(name)
	if newname != name {
		t.Errorf("Expected same name, but got %s", newname)
	}

	_, err := os.Create(name)
	if err != nil {
		t.Error("Error creating file:", err.Error())
	}

	newname = handleDuplicate(newname)
	expected := filepath.Join(home, "Downloads", "test (1).pdf")
	if newname != expected {
		t.Errorf("Expected name to be %s, but got %s", expected, newname)
	}

	_, err = os.Create(newname)
	if err != nil {
		t.Error("Error creating file:", err.Error())
	}

	newname = handleDuplicate(newname)
	expected1 := filepath.Join(home, "Downloads", "test (2).pdf")
	if newname != expected1 {
		t.Errorf("Expected name to be %s, but got %s", expected1, newname)
	}

	for _, name := range []string{name, expected} {
		if err := os.Remove(name); err != nil {
			t.Error("Error removing file:", err.Error())
		}
	}
}

func TestResumableSuccess(t *testing.T) {
	link := "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
	res, err := http.Head(link)
	if err != nil {
		t.Error("Error while fetching link:", err.Error())
	}

	isResumable := resumable(res)
	expected := true
	if isResumable != expected {
		t.Errorf("Resumable expected to be %v, but got %v", expected, isResumable)
	}
}

func TestResumableError(t *testing.T) {
	link := "https://google.com/s"
	res, err := http.Head(link)
	if err != nil {
		t.Error("Error while fetching link:", err.Error())
	}

	isResumable := resumable(res)
	expected := false
	if isResumable != expected {
		t.Errorf("Resumable expected to be %v, but got %v", expected, isResumable)
	}

	link = "https://cartographicperspectives.org/index.php/journal/article/view/cp13-full/pdf"
	res, err = http.Head(link)
	if err != nil {
		t.Error("Error while fetching link:", err.Error())
	}

	isResumable = resumable(res)
	expected = false
	if isResumable != expected {
		t.Errorf("Resumable expected to be %v, but got %v", expected, isResumable)
	}
}

func TestCalculatePartitionOneChunkLen(t *testing.T) {
	// 10 mb file, but has no header to get the desired data
	link := "https://cartographicperspectives.org/index.php/journal/article/view/cp13-full/pdf"
	entry, err := Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	t.Error(entry)
	if entry.ChunkLen() > 1 {
		t.Error("Chunk length expected to be one, but got", entry.ChunkLen())
	}
}

func TestCalculatePartitionMoreOneChunkLen(t *testing.T) {
	// 100 mb file
	link := "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=sampledocs-100mb-pdf-file&ext=pdf"
	entry, err := Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	if entry.ChunkLen() == 1 {
		t.Error("Chunk length expected to be one, but got", entry.ChunkLen())
	}

	// 1 gb file
	link = "https://bit.ly/1GB-testfile"
	entry, err = Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	if entry.ChunkLen() == 1 {
		t.Error("Chunk length expected to be one, but got", entry.ChunkLen())
	}

	// 10 mb file
	link = "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=SampleDocs-Test%20PDF%20File%20With%20Dummy%20Data%20For%20Testing&ext=pdf"
	entry, err = Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	if entry.ChunkLen() == 1 {
		t.Error("Chunk length expected to be one, but got", entry.ChunkLen())
	}

	// 50 mb file
	link = "https://link.testfile.org/PDF50MB"
	entry, err = Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	if entry.ChunkLen() == 1 {
		t.Error("Chunk length expected to be one, but got", entry.ChunkLen())
	}
}
