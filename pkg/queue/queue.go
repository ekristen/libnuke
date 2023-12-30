package queue

type IQueue interface {
	Total() int
	Count(states ...ItemState) int
}

type Queue struct {
	Items []IItem
}
