package models



type NewGroupDetails struct {
	Title string `json:"title"`
	Description string `json:"description"`
}
type Group struct {
	Id          int64    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AdminId     int64    `json:"admin_id"`
	AdminName   string `json:"admin_name"`
	MemberCount int    `json:"member_count"`
}