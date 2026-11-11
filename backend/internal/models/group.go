package models

type NewGroupDetails struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
type Group struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AdminId     int64  `json:"admin_id"`
	AdminName   string `json:"admin_name"`
	MemberCount int    `json:"member_count"`
}

type NewEvent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventAt     string `json:"event_at"`
}

type Event struct {
	Id            int64  `json:"id"`
	GroupId       int64  `json:"group_id"`
	CreatorId     int64  `json:"creator_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	EventAt       string `json:"event_at"`
	CreatedAt     string `json:"created_at"`
	GoingCount    int    `json:"going_count"`
	NotGoingCount int    `json:"not_going_count"`
	MyResponse    string `json:"my_response,omitempty"`
}

type EventVote struct {
	Response string `json:"response"`
}

type InviteRequest struct {
	UserId int64 `json:"user_id"`
}

// InvitableUser is a follower or followee of the current user,
// with their relation to a given group.
type InvitableUser struct {
	UserId    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
	Status    string `json:"status"` // "none", "invited", "pending", "member"
}
