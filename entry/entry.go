package entry

import "time"

type Entry interface {
	ID() string
	Name() string
	Location() string
	Icon() string
	Size() int64
	Type() string    // document, compressed, audio, video, image, other, etc
	URL() string     // url which the entry downloaded from
	Date() time.Time // date created
	Status() string  // on progress, pending, paused, queued, stopped
}
