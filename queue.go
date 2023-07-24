package rapid

import "log"

type (
	Queue interface {
		Push(entry Entry)
		Pop() Entry
		Len() int
		IsEmpty() bool
	}

	QueueFunc func(Setting) Queue
)

var queueMap = make(map[string]QueueFunc)

func NewQueue(provider string, setting Setting) Queue {
	queue, ok := queueMap[provider]
	if !ok {
		log.Panicf("Provider %s is not implemented", provider)
		return nil
	}

	return queue(setting)
}

func RegisterQueue(name string, queue QueueFunc) {
	queueMap[name] = queue
}
