package queue

import "sync"

type Queue struct {
	messages []string // TODO: Make generic!
	mutex    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{messages: make([]string, 0)}
}

func (q *Queue) Publish(message string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.messages = append(q.messages, message)
}

func (q *Queue) Consume() (string, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.messages) == 0 {
		return "", false
	}
	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg, true
}
