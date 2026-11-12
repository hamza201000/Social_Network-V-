"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import type { GroupChatMessage } from "@/types/chat";
import EmojiPicker from "@/components/chat/EmojiPicker";
import { API_BASE_URL, imageUrl } from "@/lib/api";
import { subscribe } from "@/lib/realtime";
import "./group-chat.css";

type GroupChatProps = {
  group_id: string;
};

function isGroupMessage(data: unknown): data is GroupChatMessage {
  return (data as { type?: string } | null)?.type === "group_message";
}

function messageTime(date: string) {
  return new Date(date).toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

// The common chat room of a group. Only members can load it; new messages
// arrive live over the shared websocket.
export default function GroupChat({ group_id }: GroupChatProps) {
  const [messages, setMessages] = useState<GroupChatMessage[]>([]);
  const [myId, setMyId] = useState<number | null>(null);
  const [draft, setDraft] = useState("");
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const listEnd = useRef<HTMLDivElement>(null);
  const input = useRef<HTMLInputElement>(null);

  function addMessage(message: GroupChatMessage) {
    setMessages((current) =>
      current.some((item) => item.id === message.id) ? current : [...current, message],
    );
  }

  useEffect(() => {
    let active = true;

    async function loadHistory() {
      try {
        const response = await fetch(
          `${API_BASE_URL}/api/groups/${encodeURIComponent(group_id)}/messages`,
          { credentials: "include" },
        );
        const data = await response.json().catch(() => null);
        if (!response.ok) throw new Error(data?.error || "Could not load the group chat.");
        if (active) {
          setMessages(data as GroupChatMessage[]);
          setError("");
        }
      } catch (loadError) {
        if (active) setError(loadError instanceof Error ? loadError.message : "Could not load the group chat.");
      } finally {
        if (active) setLoading(false);
      }
    }

    fetch(`${API_BASE_URL}/api/myprofile`, { credentials: "include" })
      .then((response) => (response.ok ? response.json() : null))
      .then((profile) => {
        if (active && profile) setMyId(profile.user_id);
      })
      .catch(() => {});

    void loadHistory();

    const unsubscribe = subscribe((data) => {
      if ((data as { type?: string })?.type === "reconnected") {
        void loadHistory();
        return;
      }
      if (isGroupMessage(data) && String(data.group_id) === String(group_id)) {
        addMessage(data);
      }
    });

    return () => {
      active = false;
      unsubscribe();
    };
  }, [group_id]);

  useEffect(() => {
    listEnd.current?.scrollIntoView({ block: "nearest" });
  }, [messages.length]);

  async function sendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const text = draft.trim();
    if (!text || sending) return;
    setSending(true);
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/groups/${encodeURIComponent(group_id)}/messages`,
        {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ message: text }),
        },
      );
      const data = await response.json().catch(() => null);
      if (!response.ok) throw new Error(data?.error || "Could not send the message.");
      addMessage(data as GroupChatMessage);
      setDraft("");
      setError("");
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : "Could not send the message.");
    } finally {
      setSending(false);
      input.current?.focus();
    }
  }

  return (
    <div className="group-chat">
      <div className="group-chat-messages" aria-live="polite">
        {loading && <p className="group-chat-status">Loading chat…</p>}
        {!loading && messages.length === 0 && !error && (
          <p className="group-chat-status">No messages yet. Start the conversation!</p>
        )}
        {messages.map((message) => {
          const mine = message.sender_id === myId;
          return (
            <div key={message.id} className={`group-chat-row${mine ? " is-mine" : ""}`}>
              {!mine && (
                <span className="group-chat-avatar" aria-hidden="true">
                  {message.avatar_sender ? (
                    <img src={imageUrl(message.avatar_sender)} alt="" />
                  ) : (
                    message.username_sender.slice(0, 1).toUpperCase()
                  )}
                </span>
              )}
              <div className="group-chat-content">
                {!mine && <span className="group-chat-sender">{message.username_sender}</span>}
                <p className="group-chat-bubble">{message.message}</p>
                <time className="group-chat-time">{messageTime(message.created_at)}</time>
              </div>
            </div>
          );
        })}
        <div ref={listEnd} />
      </div>

      {error && <p className="group-chat-error" role="alert">{error}</p>}

      <form className="group-chat-composer" onSubmit={sendMessage}>
        <EmojiPicker
          onPick={(emoji) => {
            setDraft((current) => current + emoji);
            input.current?.focus();
          }}
        />
        <input
          ref={input}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          placeholder="Message the group…"
          aria-label="Write a message to the group"
          maxLength={5000}
        />
        <button type="submit" disabled={!draft.trim() || sending}>
          Send
        </button>
      </form>
    </div>
  );
}
