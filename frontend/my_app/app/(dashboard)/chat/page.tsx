"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import ConversationList from "@/components/chat/conv_list";
import type { ChatMessage, ChatUser, NewMessage } from "@/types/chat";
import "./chat.css";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function websocketUrl() {
  const url = new URL(API_URL);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = "/api/ws";
  return url.toString();
}

function initials(username: string) {
  return username.slice(0, 2).toUpperCase();
}

function messageTime(date: string) {
  return new Date(date).toLocaleTimeString([], {
    hour: "numeric",
    minute: "2-digit",
  });
}

async function checkResponse(response: Response) {
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.error ?? "Chat request failed");
  }
}

async function markMessageSeen(messageId: number) {
  const response = await fetch(`${API_URL}/api/messages/${messageId}/seen`, {
    method: "PUT",
    credentials: "include",
  });
  await checkResponse(response);
}

export default function ChatPage() {
  const [conversations, setConversations] = useState<ChatUser[]>([]);
  const [selectedUser, setSelectedUser] = useState<ChatUser | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [error, setError] = useState("");
  const [mobileChatOpen, setMobileChatOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const socket = useRef<WebSocket | null>(null);
  const selectedUserRef = useRef<ChatUser | null>(null);

  selectedUserRef.current = selectedUser;

  useEffect(() => {
    let open = true;

    async function loadConversations() {
      try {
        const response = await fetch(`${API_URL}/api/chats`, {
          credentials: "include",
        });
        await checkResponse(response);
        const users: ChatUser[] = await response.json();
        if (open) setConversations(users);
      } catch (requestError) {
        if (open) {
          setError(
            requestError instanceof Error
              ? requestError.message
              : "Could not load chats",
          );
        }
      } finally {
        if (open) setLoading(false);
      }
    }

    function connectWebSocket() {
      const connection = new WebSocket(websocketUrl());
      socket.current = connection;

      connection.onopen = () => {
        void loadConversations();
      };
      connection.onmessage = (event) => {
        const message: ChatMessage = JSON.parse(event.data);
        if (message.type === "error") {
          setError(message.message);
          return;
        }
        setError("");

        const currentChat = selectedUserRef.current;
        if (
          currentChat &&
          (message.sender_id === currentChat.user_id ||
            message.receiver_id === currentChat.user_id)
        ) {
          setMessages((current) =>
            current.some((item) => item.id === message.id)
              ? current
              : [...current, message],
          );

          if (message.sender_id === currentChat.user_id && !message.seen) {
            void markMessageSeen(message.id)
              .then(() =>
                setMessages((current) =>
                  current.map((item) =>
                    item.id === message.id ? { ...item, seen: true } : item,
                  ),
                ),
              )
              .catch(() =>
                setError("Could not mark the message as seen"),
              );
          }
        }

        void loadConversations();
      };
      connection.onerror = () => {
        setError("Could not connect to chat. Check that the backend is running.");
      };
    }

    void loadConversations();
    connectWebSocket();

    return () => {
      open = false;
      socket.current?.close();
      socket.current = null;
    };
  }, []);

  async function markAsSeen(messageId: number) {
    try {
      await markMessageSeen(messageId);
      setMessages((current) =>
        current.map((message) =>
          message.id === messageId ? { ...message, seen: true } : message,
        ),
      );
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "Could not mark the message as seen",
      );
    }
  }

  async function selectConversation(user: ChatUser) {
    setSelectedUser(user);
    setMessages([]);
    setMobileChatOpen(true);
    setError("");

    try {
      const response = await fetch(
        `${API_URL}/api/chats/${user.user_id}/messages`,
        { credentials: "include" },
      );
      await checkResponse(response);
      const history: ChatMessage[] = await response.json();
      setMessages(history);

      for (const message of history) {
        if (message.receiver_id !== user.user_id && !message.seen) {
          await markAsSeen(message.id);
        }
      }
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "Could not load messages",
      );
    }
  }

  async function sendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const text = draft.trim();
    if (!text || !selectedUser) return;

    const request: NewMessage = {
      receiver_id: selectedUser.user_id,
      message: text,
    };

    if (socket.current?.readyState === WebSocket.OPEN) {
      socket.current.send(JSON.stringify(request));
      setDraft("");
      return;
    }

    try {
      const response = await fetch(`${API_URL}/api/messages`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
      });
      await checkResponse(response);
      const message: ChatMessage = await response.json();
      setMessages((current) => [...current, message]);
      setConversations((current) =>
        current.map((user) =>
          user.user_id === selectedUser.user_id
            ? { ...user, last_msg: message }
            : user,
        ),
      );
      setDraft("");
      setError("");
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : "Could not send the message",
      );
    }
  }

  return (
    <main className="chat-page">
      <section
        className={`chat-shell${mobileChatOpen ? " mobile-chat-open" : ""}`}
        aria-label="Messages"
      >
        <aside className="conversation-sidebar">
          <header className="inbox-header">
            <div>
              <span className="eyebrow">YOUR SPACE</span>
              <h1>
                Messages<span className="inbox-count">{conversations.length}</span>
              </h1>
            </div>
          </header>

          <ConversationList
            conversations={conversations}
            selectedUserId={selectedUser?.user_id ?? null}
            onSelect={selectConversation}
          />

          <footer className="sidebar-footer">
            {loading ? "Loading chats..." : `${conversations.length} conversations`}
          </footer>
        </aside>

        <section className="chat-conversation" aria-label="Conversation">
          {selectedUser ? (
            <>
              <header className="chat-header">
                <button
                  className="icon-button mobile-back"
                  type="button"
                  onClick={() => setMobileChatOpen(false)}
                  aria-label="Back to conversations"
                >
                  ‹
                </button>
                <span className="chat-avatar tone-green">
                  {initials(selectedUser.username)}
                  {selectedUser.is_online && <i className="avatar-presence" />}
                </span>
                <div className="chat-header-copy">
                  <h2>{selectedUser.username}</h2>
                  <p>
                    <span
                      className={
                        selectedUser.is_online ? "online-dot" : "status-dot"
                      }
                    />
                    {selectedUser.is_online ? "Online" : "Offline"}
                  </p>
                </div>
              </header>

              <div className="message-area" aria-live="polite">
                <div className="message-stack">
                  {messages.map((message) => {
                    const fromMe = message.sender_id !== selectedUser.user_id;
                    return (
                      <div
                        key={message.id}
                        className={`message-row${fromMe ? " message-row-mine" : ""}`}
                      >
                        <div
                          className={`message-content${fromMe ? " message-content-mine" : ""}`}
                        >
                          <p className="message-bubble">{message.message}</p>
                          <time className="message-time">
                            {messageTime(message.created_at)}
                            {fromMe && (
                              <span className="read-check">
                                {message.seen ? " ✓✓" : " ✓"}
                              </span>
                            )}
                          </time>
                        </div>
                      </div>
                    );
                  })}
                  {messages.length === 0 && (
                    <p className="empty-search">No messages yet. Say hello!</p>
                  )}
                </div>
              </div>

              <form className="message-composer" onSubmit={sendMessage}>
                <input
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder={`Message ${selectedUser.username}...`}
                  aria-label="Write a message"
                />
                <button
                  className="send-button"
                  type="submit"
                  disabled={!draft.trim()}
                >
                  Send
                </button>
              </form>
            </>
          ) : (
            <div className="chat-empty-state">
              <h2>No conversation selected</h2>
              <p>Choose a person from your chats to view messages.</p>
            </div>
          )}
          {error && <p className="chat-error" role="alert">{error}</p>}
        </section>
      </section>
    </main>
  );
}
