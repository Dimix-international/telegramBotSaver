package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(e Event) error
}

type Type int

const (
	Unknown Type = iota //для объединения группы констант, для первой будет равной 0, для второй - 1
	Message             //1
)

type Event struct {
	Type Type
	Text string
	Meta interface{} //нужен для разных конкретных реализаций, т.к. Event структура глобальная для всех
}