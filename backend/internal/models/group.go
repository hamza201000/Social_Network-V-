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
	// MyStatus is the viewer's relation to the group:
	// "admin", "member", "invited", "pending" or "none".
	MyStatus string `json:"my_status"`
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

// GroupMember is a member of a group, or a user with a pending join request.
type GroupMember struct {
	UserId    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
	Role      string `json:"role,omitempty"` // "admin" or "member"
	CreatedAt string `json:"created_at"`
}

// DecisionRequest is the body used to accept or refuse an invitation or a join request.
type DecisionRequest struct {
	Accept bool `json:"accept"`
}

// GroupUpdatePush is sent over the websocket to tell open pages that a
// group's members, requests or invitations changed. It is not stored.
type GroupUpdatePush struct {
	Type    string `json:"type"` // always "group_update"
	GroupId int64  `json:"group_id"`
	Action  string `json:"action"`
	UserId  int64  `json:"user_id"`
}
