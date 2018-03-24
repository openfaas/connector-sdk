package types

import "sync"

func NewTopicMap() TopicMap {
	lookup := make(map[string][]string)
	return TopicMap{
		lookup: &lookup,
		lock:   sync.Mutex{},
	}
}

type TopicMap struct {
	lookup *map[string][]string
	lock   sync.Mutex
}

func (t *TopicMap) Match(topicName string) []string {
	t.lock.Lock()

	var values []string

	for key, val := range *t.lookup {
		if key == topicName {
			values = val
			break
		}
	}

	t.lock.Unlock()

	return values
}

func (t *TopicMap) Sync(updated *map[string][]string) {
	t.lock.Lock()

	t.lookup = updated

	t.lock.Unlock()
}
