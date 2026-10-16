import type { ChatUser } from "@/types/chat";

interface ConversationListProps {
  conversations: ChatUser[];
  selectedUserId: number | null;
  onSelect: (user: ChatUser) => void;
}

function initials(username: string) {
  return username.slice(0, 2).toUpperCase();
}

function messageTime(date: string) {
  if (!date) return "";
  return new Date(date).toLocaleTimeString([], {
    hour: "numeric",
    minute: "2-digit",
  });
}

export default function ConversationList({
  conversations,
  selectedUserId,
  onSelect,
}: ConversationListProps) {
  return (
    <div className="conversation-list">
      <div className="list-label">RECENT</div>
      {conversations.length === 0 ? (
        <p className="empty-search">No conversations yet.</p>
      ) : (
        conversations.map((user) => (
          <button
            key={user.user_id}
            type="button"
            className={`conversation-item${selectedUserId === user.user_id ? " conversation-active" : ""}`}
            onClick={() => onSelect(user)}
          >
            <span
              className="chat-avatar tone-green"
              aria-label={user.username}
            >
              {initials(user.username)}
              {user.is_online && <i className="avatar-presence" />}
            </span>
            <span className="conversation-copy">
              <span className="conversation-name-row">
                <strong>{user.username}</strong>
                <time>{messageTime(user.last_msg?.created_at ?? "")}</time>
              </span>
              <span className="conversation-preview-row">
                <span className="conversation-type">
                  {user.last_msg?.message ?? "Start a conversation"}
                </span>
              </span>
            </span>
          </button>
        ))
      )}
    </div>
  );
}
