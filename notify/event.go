package notify

import (
	"fmt"
	"strings"
	"time"
)

// Event — «что-то случилось» в системе. Единственная форма, которую видят
// правила: `<Entity>.<Action>` («order.created», «response.created») плюс
// снапшот сущности в `Payload` (по нему сверяются условия и рендерятся шаблоны).
//
// Публикуется из CRUD-хуков (`baseR`) либо вручную из ручки. Bus разносит
// событие подписчикам, Dispatcher матчит его с правилами из БД.
type Event struct {
	// Entity — имя сущности в kebab-case, как у роутов («order», «event»).
	Entity string
	// Action — created | updated | deleted | <кастомное>.
	Action string
	// ActorID — кто инициировал (id пользователя), если известно.
	ActorID string
	// ItemID — id сущности.
	ItemID string
	// Payload — плоский снапшот полей сущности (name, status, сумма, …).
	Payload map[string]any
	// Data — необязательный богатый объект для шаблонов правил (полный заказ,
	// отзыв, …). Доступен в шаблоне как `{{.data.Field}}`; nil — не используется.
	Data any
	// At — момент события.
	At time.Time
}

// Name — полное имя события «entity.action».
func (e Event) Name() string {
	return e.Entity + "." + e.Action
}

// Match — подходит ли событие шаблону правила: «*», «*.*», «order.*», «order.created».
// Пустой шаблон тоже считается совпадением (правило «на всё»).
func (e Event) Match(pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" || pattern == "*.*" {
		return true
	}
	parts := strings.SplitN(pattern, ".", 2)
	entity, action := parts[0], "*"
	if len(parts) == 2 {
		action = parts[1]
	}
	if entity != "*" && entity != e.Entity {
		return false
	}
	if action != "*" && action != e.Action {
		return false
	}
	return true
}

// flat — плоское представление события для условий и шаблонов.
func (e Event) flat() map[string]string {
	out := make(map[string]string, len(e.Payload)+4)
	out["entity"] = e.Entity
	out["action"] = e.Action
	out["id"] = e.ItemID
	out["actorId"] = e.ActorID
	for k, v := range e.Payload {
		out[k] = fmt.Sprint(v)
	}
	return out
}

// Message — готовое к отправке сообщение (результат матчинга правила).
// Канал получает уже отрендеренные To/Subject/Body и не знает про правила.
type Message struct {
	Channel string
	To      string
	Subject string
	Body    string
	Meta    map[string]string
}
