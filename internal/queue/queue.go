package queue

import (
	"fmt"
	"sync"
)

type QueueManager struct {
	queues map[string]*Queue
	nextId int
	mutex  sync.Mutex
}

type Queue struct {
	id       int
	name     string
	messages []string // TODO: Make generic!
	mutex    sync.Mutex
}

func NewQueueManager() *QueueManager {
	qm := &QueueManager{
		queues: make(map[string]*Queue),
		nextId: 1,
	}

	// Create default queue
	qm.queues["default"] = NewQueue("default", 0)
	return qm
}

func (qm *QueueManager) CreateQueue(name string) (int, error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	if _, exists := qm.queues[name]; exists {
		return 0, fmt.Errorf("queue with name '%s' already exists", name)
	}

	id := qm.nextId
	qm.queues[name] = NewQueue(name, qm.nextId)
	qm.nextId++
	return id, nil
}

func (qm *QueueManager) ListQueues() []string {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	names := make([]string, 0, len(qm.queues))
	for name := range qm.queues {
		names = append(names, name)
	}
	return names
}

func (qm *QueueManager) GetQueue(name string) (*Queue, bool) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	queue, exists := qm.queues[name]
	return queue, exists
}

func NewQueue(name string, id int) *Queue {
	return &Queue{id: id, name: name, messages: make([]string, 0)}
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
