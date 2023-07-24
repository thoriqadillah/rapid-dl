package rapid

import "io"

type progressBar struct {
	Entry
	onprogress OnProgress
	reader     io.ReadCloser
	index      int
	downloaded int64
	progress   float64
	chunkSize  int64
}

func (r *progressBar) Read(payload []byte) (n int, err error) {
	n, err = r.reader.Read(payload)
	if err != nil {
		return n, err
	}

	r.downloaded += int64(n)
	r.progress = float64(100 * r.downloaded / r.chunkSize)

	if r.onprogress != nil {
		r.onprogress(
			r.ID(),
			r.index,
			r.downloaded,
			r.progress,
		)
	}

	return n, err
}

func (r *progressBar) Close() error {
	return r.reader.Close()
}
