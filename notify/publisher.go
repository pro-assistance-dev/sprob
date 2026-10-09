package notify

import "context"

// PublishNotify — реализация `helper.EventPublisher`: удобная форма публикации
// события из кода (без сборки Event руками). Payload — плоский снапшот сущности,
// `data` — необязательный богатый объект для шаблонов (`{{.data.Field}}`).
func (n *Notifier) PublishNotify(entity, action, itemID, actorID string, payload map[string]any, data any) {
	if n == nil {
		return
	}
	n.Publish(context.Background(), Event{
		Entity:  entity,
		Action:  action,
		ItemID:  itemID,
		ActorID: actorID,
		Payload: payload,
		Data:    data,
	})
}

// PublishData — публикация с богатым объектом для шаблонов (`{{.data.Field}}`).
func (n *Notifier) PublishData(entity, action, itemID string, payload map[string]any, data any) {
	if n == nil {
		return
	}
	n.Publish(context.Background(), Event{
		Entity:  entity,
		Action:  action,
		ItemID:  itemID,
		Payload: payload,
		Data:    data,
	})
}
