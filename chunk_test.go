package rapid

import (
	"testing"
)

func TestChunkRangeOneChunkLen(t *testing.T) {
	link := "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
	entry, err := Fetch(link, DefaultSetting())
	if err != nil {
		t.Error("Error fetching url:", err.Error())
	}

	chunkSize := entry.Size() / int64(entry.ChunkLen())
	var start int64 = 0
	var end int64 = 0
	for i := 0; i < entry.ChunkLen(); i++ {
		start, end = calculatePosition(entry, chunkSize, i)
	}

	if start != 0 {
		t.Errorf("Start range expected to be 0, but got %d", start)
	}

	if end != entry.Size() {
		t.Errorf("Start range expected to be 0, but got %d", start)
	}
}
