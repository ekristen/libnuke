package queue

type IQueue interface {
	Total() int
	Count(states ...ItemState) int
}

type Queue struct {
	Items []*Item
}

func (q Queue) GetItems() []*Item {
	return q.Items
}

func (q Queue) Total() int {
	return len(q.Items)
}

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
