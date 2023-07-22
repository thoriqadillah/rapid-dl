package rapid

type (
	List interface {
		Push(entry Entry)
		Pop() Entry
		Len() int
		IsEmpty() bool
		Get(i int) Entry
	}

	queue struct {
		entries []Entry
	}
)

func NewQueue() List {
	return &queue{
		entries: make([]Entry, 0),
	}
}

func (q *queue) Push(entry Entry) {
	q.entries = append(q.entries, entry)
}

func (q *queue) Pop() Entry {
	first := q.entries[0]
	q.entries = q.entries[1:]

	return first
}

func (q *queue) Len() int {
	return len(q.entries)
}

func (q *queue) IsEmpty() bool {
	return q.Len() == 0
}

func (q *queue) Get(i int) Entry {
	return q.entries[i]
}
