// Package queue provides a simple list mechanism with some helper functions to determine current counts based on
// resource type or state.
package queue

type IQueue interface {
	Total() int
	Count(states ...ItemState) int
}

// Queue provides a very simple interface for queuing Item for processing
type Queue struct {
	Items []*Item
}

// GetItems returns all the items currently in the Queue
func (q Queue) GetItems() []*Item {
	return q.Items
}

// Total returns the total number of items in the Queue
func (q Queue) Total() int {
	return len(q.Items)
}

// Count returns the total number of items in a specific ItemState from the Queue
func (q Queue) Count(states ...ItemState) int {
	count := 0
	for _, item := range q.Items {
		for _, state := range states {
			if item.GetState() == state {
				count++
				break
			}
		}
	}
	return count
}

// CountByType returns the total number of items that match a ResourceType and specific ItemState from the Queue
func (q Queue) CountByType(resourceType string, states ...ItemState) int {
	count := 0
	for _, item := range q.Items {
		if item.Type == resourceType {
			for _, state := range states {
				if item.GetState() == state {
					count++
					break
				}
			}
		}
	}
	return count
}
