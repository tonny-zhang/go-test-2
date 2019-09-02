package main

type Event struct {
	listeners map[string][]func(argv ...interface{})
}

type EventSub struct {
	Event
}

func (event *Event) on(eventName string, listener func(argv ...interface{})) {
	if event.listeners == nil {
		event.listeners = make(map[string][]func(argv ...interface{}))
	}

	event.listeners[eventName] = append(event.listeners[eventName], listener)
}

func (event *Event) emit(eventName string, argv ...interface{}) {
	arr, isExists := event.listeners[eventName]
	if isExists {
		for _, listener := range arr {
			listener(argv...)
		}
	}
}
func (event *Event) off(eventName string) {
	delete(event.listeners, eventName)
}
