package models

// Notification types.
const (
	NotifNewFollower = "new_follower"
	NotifGroupInvite = "group_invite"
	NotifJoinRequest = "join_request"
	NotifGroupEvent  = "group_event"
)

type Notification struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	ActorID       int64  `json:"actor_id"`
	ActorUsername string `json:"actor_username"`
	ActorAvatar   string `json:"actor_avatar"`
	Type          string `json:"type"`
	GroupID       *int64 `json:"group_id"`
	GroupTitle    string `json:"group_title"`
	EventID       *int64 `json:"event_id"`
	Message       string `json:"message"`
	IsRead        bool   `json:"is_read"`
	CreatedAt     string `json:"created_at"`
}

// NotificationList is the response of GET /api/notifications.
type NotificationList struct {
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unread_count"`
}

// NotificationPush is what is sent over the websocket when a notification is created.
// The "type" field is always "notification" so the client can tell it apart from chat messages.
type NotificationPush struct {
	Type         string       `json:"type"`
	Notification Notification `json:"notification"`
}
