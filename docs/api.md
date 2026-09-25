## Posts

All endpoints use the existing `session_id` HttpOnly cookie. `GET` endpoints may be used without a session, but only public posts are returned to anonymous users.

### Create a post

`POST /api/posts`

Use JSON when no image is needed:

```json
{
	"content": "Hello",
	"privacy": "public",
	"follower_ids": [12, 18]
}
```

Use `multipart/form-data` to include an image. Fields are `content`, `privacy`, `follower_ids` (comma-separated user IDs), and `image`. Accepted image types are JPEG, PNG, WebP, and GIF, up to 5 MB.

`privacy` must be `public`, `almost_private`, or `private`. A private post must include at least one selected follower, and every selected user must follow the author.

### List posts

`GET /api/posts`

Returns public posts plus posts visible to the current session.

### Get one post

`GET /api/posts/{id}`

Returns `404` when the post does not exist or is not visible to the current session.

## Comments

### Create a comment

`POST /api/posts/{id}/comments`

Use JSON with `content`, or `multipart/form-data` with `content` and an optional `image`. The same image limits apply. The user must be logged in and able to view the post.

### List comments

`GET /api/posts/{id}/comments`

Returns comments only when the current session can view the parent post.

Successful create responses use HTTP `201` and return the new ID in `data.id`. Invalid input returns `400`; missing or expired sessions return `401`.
