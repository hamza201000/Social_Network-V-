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

## Private chats

Chat endpoints require the existing `session_id` HttpOnly cookie. The logged-in user is always taken from that session; the request cannot choose its own sender ID. A conversation can be messaged when either user follows the other.

### Send a message

`POST /api/messages`

```json
{
  "receiver_id": 12,
  "message": "Hi there"
}
```

The message is stored even when the receiver is offline. A `403` response means neither user follows the other.

### Get a conversation

`GET /api/chats/{userID}/messages`

Returns messages between the logged-in user and the user identified in the path, ordered from oldest to newest.

### List chats

`GET /api/chats`

Returns users the logged-in user has exchanged messages with, including the latest message and whether the user currently has an open WebSocket connection.

### Mark a message as seen

`PUT /api/messages/{id}/seen`

Only the message receiver can mark it as seen. Repeating this request is safe.

### Receive messages over WebSocket

`GET /api/ws`

Open this endpoint with the logged-in user's `session_id` cookie. Send JSON frames with `receiver_id` and `message` fields, the same format used by `POST /api/messages`. The saved message is sent to the receiver immediately if they are connected and echoed back to the sending WebSocket. Messages sent over HTTP are also delivered to connected receivers.

Example browser console (after logging in on the same site):

```js
const socket = new WebSocket("ws://localhost:8080/api/ws");
socket.onmessage = (event) => console.log(JSON.parse(event.data));
socket.onopen = () => socket.send(JSON.stringify({
  receiver_id: 12,
  message: "Hello over WebSocket"
}));
```

The WebSocket hub keeps presence in memory, so online status is local to the running server and resets when the server restarts.

### Run locally and try the API

From the repository root, start the backend (database migrations run automatically):

```powershell
cd backend
go mod download
go run cmd/server/main.go
```

Log in first and save the returned session cookie in `cookies.txt`, then send a message and retrieve the conversation:

```powershell
curl.exe -b cookies.txt -H "Content-Type: application/json" -d "{\"receiver_id\":12,\"message\":\"Hi there\"}" http://localhost:8080/api/messages
curl.exe -b cookies.txt http://localhost:8080/api/chats/12/messages
curl.exe -b cookies.txt http://localhost:8080/api/chats
curl.exe -X PUT -b cookies.txt http://localhost:8080/api/messages/1/seen
```
