package rapid

type defaultQueue struct {
	entries []Entry
}

const QueueDefault = "default"

func newDefaultQueue(setting Setting) Queue {
	return &defaultQueue{
		entries: make([]Entry, 0),
	}
}

func (q *defaultQueue) Push(entry Entry) {
	q.entries = append(q.entries, entry)
}

func (q *defaultQueue) Pop() Entry {
	first := q.entries[0]
	q.entries = q.entries[1:]

	return first
}

func (q *defaultQueue) Len() int {
	return len(q.entries)
}

func (q *defaultQueue) IsEmpty() bool {
	return q.Len() == 0
}

func init() {
	RegisterQueue(QueueDefault, newDefaultQueue)
}
