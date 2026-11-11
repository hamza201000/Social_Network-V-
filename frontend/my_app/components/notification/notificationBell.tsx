"use client";

import { API_BASE_URL, imageUrl } from "@/lib/api";
import {
  AppNotification,
  NotificationList,
  NotificationPush,
} from "@/types/notification";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import "./notification.css";

function websocketUrl() {
  const url = new URL(API_BASE_URL);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = "/api/ws";
  return url.toString();
}

// SQLite returns "2026-11-11 16:21:23" (UTC, no zone) or an RFC3339 string
function parseDate(value: string) {
  const iso = value.includes("T") ? value : value.replace(" ", "T");
  return new Date(/Z|[+-]\d\d:?\d\d$/.test(iso) ? iso : `${iso}Z`);
}

function timeAgo(value: string) {
  const seconds = Math.max(0, (Date.now() - parseDate(value).getTime()) / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return parseDate(value).toLocaleDateString();
}

// where clicking a notification takes the user
function notificationLink(n: AppNotification) {
  if (n.type === "new_follower") return `/profile/${encodeURIComponent(n.actor_username)}`;
  if (n.group_id) return `/group/${n.group_id}`;
  return null;
}

const ICONS: Record<AppNotification["type"], string> = {
  new_follower: "👤",
  group_invite: "✉️",
  join_request: "🙋",
  group_event: "📅",
  invite_accepted: "🤝",
  invite_declined: "✖️",
  join_accepted: "✅",
  join_refused: "⛔",
};

export default function NotificationBell() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState<AppNotification[]>([]);
  const [unread, setUnread] = useState(0);
  const [error, setError] = useState("");
  const [ring, setRing] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  // initial load + live updates through the websocket
  useEffect(() => {
    let active = true;
    let socket: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    let retryDelay = 1000;

    async function load() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/notifications`, {
          credentials: "include",
        });
        if (!response.ok) throw new Error();
        const data: NotificationList = await response.json();
        if (!active) return;
        setNotifications(data.notifications ?? []);
        setUnread(data.unread_count ?? 0);
        setError("");
      } catch {
        if (active) setError("Could not load notifications.");
      }
    }

    function connect() {
      socket = new WebSocket(websocketUrl());

      socket.onopen = () => {
        retryDelay = 1000;
        // reload in case something arrived while we were disconnected
        void load();
      };

      socket.onmessage = (event) => {
        let data: unknown;
        try {
          data = JSON.parse(event.data);
        } catch {
          return;
        }
        // the same socket also carries chat messages: ignore everything else
        if ((data as { type?: string })?.type !== "notification") return;
        const { notification } = data as NotificationPush;
        setNotifications((current) =>
          current.some((n) => n.id === notification.id)
            ? current
            : [notification, ...current].slice(0, 50),
        );
        setUnread((count) => count + 1);
        setRing(true);
      };

      socket.onclose = () => {
        if (!active) return;
        // reconnect with backoff (1s, 2s, 4s ... max 30s)
        retryTimer = setTimeout(connect, retryDelay);
        retryDelay = Math.min(retryDelay * 2, 30000);
      };
    }

    void load();
    connect();

    return () => {
      active = false;
      clearTimeout(retryTimer);
      socket?.close();
    };
  }, []);

  // stop the bell animation after it played
  useEffect(() => {
    if (!ring) return;
    const timer = setTimeout(() => setRing(false), 900);
    return () => clearTimeout(timer);
  }, [ring]);

  // close the dropdown on outside click or Escape
  useEffect(() => {
    if (!open) return;
    function onClick(event: MouseEvent) {
      if (!wrapperRef.current?.contains(event.target as Node)) setOpen(false);
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onClick);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onClick);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  async function markRead(n: AppNotification) {
    if (n.is_read) return;
    setNotifications((current) =>
      current.map((item) => (item.id === n.id ? { ...item, is_read: true } : item)),
    );
    setUnread((count) => Math.max(0, count - 1));
    try {
      await fetch(`${API_BASE_URL}/api/notifications/${n.id}/read`, {
        method: "PUT",
        credentials: "include",
      });
    } catch {
      // not critical: it will show as unread again on the next load
    }
  }

  async function markAllRead() {
    setNotifications((current) => current.map((n) => ({ ...n, is_read: true })));
    setUnread(0);
    try {
      await fetch(`${API_BASE_URL}/api/notifications/read-all`, {
        method: "PUT",
        credentials: "include",
      });
    } catch {
      setError("Could not update notifications.");
    }
  }

  function handleClick(n: AppNotification) {
    void markRead(n);
    const link = notificationLink(n);
    if (link) {
      setOpen(false);
      router.push(link);
    }
  }

  return (
    <div className="notif-wrapper" ref={wrapperRef}>
      <button
        type="button"
        className={`notif-bell ${ring ? "notif-bell-ring" : ""}`}
        aria-label={unread > 0 ? `Notifications, ${unread} unread` : "Notifications"}
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        <span aria-hidden="true">🔔</span>
        {unread > 0 && (
          <span className="notif-badge">{unread > 99 ? "99+" : unread}</span>
        )}
      </button>

      {open && (
        <div className="notif-panel" role="dialog" aria-label="Notifications">
          <div className="notif-header">
            <h2>Notifications</h2>
            {unread > 0 && (
              <button type="button" className="notif-mark-all" onClick={markAllRead}>
                Mark all as read
              </button>
            )}
          </div>

          {error && <p className="notif-message notif-error">{error}</p>}

          {notifications.length === 0 && !error ? (
            <p className="notif-message">You have no notifications yet.</p>
          ) : (
            <ul className="notif-list">
              {notifications.map((n) => (
                <li key={n.id}>
                  <button
                    type="button"
                    className={`notif-item ${n.is_read ? "" : "notif-unread"}`}
                    onClick={() => handleClick(n)}
                  >
                    <span className="notif-avatar">
                      {n.actor_avatar ? (
                        <img src={imageUrl(n.actor_avatar)} alt="" />
                      ) : (
                        n.actor_username.slice(0, 1).toUpperCase()
                      )}
                      <span className="notif-type-icon" aria-hidden="true">
                        {ICONS[n.type] ?? "🔔"}
                      </span>
                    </span>
                    <span className="notif-copy">
                      <span className="notif-text">
                        <strong>@{n.actor_username}</strong> {n.message}
                      </span>
                      <span className="notif-time">{timeAgo(n.created_at)}</span>
                    </span>
                    {!n.is_read && <span className="notif-dot" aria-label="unread" />}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
