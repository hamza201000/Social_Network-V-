import { API_BASE_URL } from "@/lib/api";

// One shared websocket for components that only need to listen to server
// pushes (group updates, ...). It connects on the first subscriber,
// reconnects with backoff, and closes when the last subscriber leaves.

type Listener = (data: unknown) => void;

const listeners = new Set<Listener>();
let socket: WebSocket | null = null;
let retryTimer: ReturnType<typeof setTimeout> | undefined;
let retryDelay = 1000;

function websocketUrl() {
  const url = new URL(API_BASE_URL);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = "/api/ws";
  return url.toString();
}

function connect() {
  if (socket || listeners.size === 0) return;
  const ws = new WebSocket(websocketUrl());
  socket = ws;

  ws.onopen = () => {
    retryDelay = 1000;
    // let listeners refresh anything they may have missed while offline
    listeners.forEach((listener) => listener({ type: "reconnected" }));
  };

  ws.onmessage = (event) => {
    let data: unknown;
    try {
      data = JSON.parse(event.data);
    } catch {
      return;
    }
    listeners.forEach((listener) => listener(data));
  };

  ws.onclose = () => {
    if (socket === ws) socket = null;
    if (listeners.size === 0) return;
    retryTimer = setTimeout(connect, retryDelay);
    retryDelay = Math.min(retryDelay * 2, 30000);
  };
}

export function subscribe(listener: Listener): () => void {
  listeners.add(listener);
  connect();
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) {
      clearTimeout(retryTimer);
      socket?.close();
      socket = null;
    }
  };
}

export type GroupUpdate = {
  type: "group_update";
  group_id: number;
  action: "invited" | "join_requested" | "member_joined" | "request_refused" | "invite_declined";
  user_id: number;
};

export function isGroupUpdate(data: unknown): data is GroupUpdate {
  return (data as { type?: string } | null)?.type === "group_update";
}
