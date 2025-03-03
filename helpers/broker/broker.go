package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// the amount of time to wait when pushing a message to
// a slow client or a client that closed after `range clients` started.
const patience time.Duration = time.Minute * 60

type notificationEvent struct {
	EventName string
	Payload   interface{}
}

type notifierChan chan notificationEvent

type Broker struct {
	notifier       notifierChan
	newClients     chan notifierChan
	closingClients chan notifierChan
	clients        map[notifierChan]bool
}

func NewBroker() (broker *Broker) {
	b := &Broker{
		notifier:       make(notifierChan, 50000),
		newClients:     make(chan notifierChan),
		closingClients: make(chan notifierChan),
		clients:        make(map[notifierChan]bool),
	}
	go b.Listen()
	return b
}

func (broker *Broker) SendEvent(eventName string, item interface{}) {
	event := notificationEvent{Payload: item, EventName: eventName}
	broker.notifier <- event
}

func (broker *Broker) ServeHTTP(c *gin.Context) {
	eventName := c.Param("channel")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	messageChan := make(notifierChan)
	broker.newClients <- messageChan

	defer func() {
		broker.closingClients <- messageChan
		close(messageChan)
		messageChan = nil
	}()

	notify := c.Writer.(http.CloseNotifier).CloseNotify()
	w := c.Writer
	// notify := c.Request.Context().Done()
	f, ok := w.(http.Flusher)

	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("streaming unsupported"))
		return
	}

	for {
		select {
		case <-notify:
			return
		case <-c.Request.Context().Done():
			// remove this client from the map of connected clients
			broker.closingClients <- messageChan
			return
		default:
			event := <-messageChan
			switch eventName {
			case event.EventName:
				payload, err := json.Marshal(event.Payload)
				if err != nil {
					_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("wrong json"))
				}
				fmt.Fprintf(w, "data: %s\n\n", payload)
				f.Flush()
			}
		}
	}

	//c.Stream(func(w io.Writer) bool {
	//	event := <-messageChan
	//	switch eventName {
	//	case event.EventName:
	//		payload, err := json.Marshal(event.Payload)
	//		if err != nil {
	//			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("wrong json"))
	//		}
	//		fmt.Fprintf(w, "data: %s\n\n", payload)
	//		f.Flush()
	//	}
	//	return true
	//})
}

// Listen for new notifications and redistribute them to clients
func (broker *Broker) Listen() {
	for {
		select {
		case s := <-broker.newClients:
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.notifier:
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Print("Skipping client.")
				}
			}
		}
	}
}
