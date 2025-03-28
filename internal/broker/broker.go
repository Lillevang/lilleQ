package broker

import (
	"fmt"
	"sync"
)

type Broker struct {
	topics map[string]*Topic
	mutex  sync.Mutex
}

type Topic struct {
	name        string
	subscribers []chan string
	mutex       sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*Topic),
	}
}

// Create new topic
func (b *Broker) CreateTopic(name string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.topics[name]; exists {
		return fmt.Errorf("topic with name '%s' already exists", name)
	}

	b.topics[name] = &Topic{
		name:        name,
		subscribers: make([]chan string, 0),
	}
	return nil
}

// Get a topic
func (b *Broker) GetTopic(name string) (*Topic, bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	topic, exists := b.topics[name]
	return topic, exists
}

// Subscribe to a topic
func (t *Topic) Subscribe() chan string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	subscriber := make(chan string)
	t.subscribers = append(t.subscribers, subscriber)
	return subscriber
}

// Publish a message to a topic
func (t *Topic) Publish(message string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, subscriber := range t.subscribers {
		go func(sub chan string) {
			sub <- message
		}(subscriber)
	}
}
