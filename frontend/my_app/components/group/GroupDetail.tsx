"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { Group, GroupMember } from "@/types/groupData";
import type { GroupEvent } from "@/types/eventData";
import "./group-detail.css";
import CreatePostGp from "../post/createPostGp";
import Post from "../post/post";
import CreateEvent from "@/components/events/createEvent";
import GetEvents from "@/components/events/GetEvents";
import GroupChat from "./GroupChat";
import { API_BASE_URL, imageUrl } from "@/lib/api";
import { isGroupUpdate, subscribe } from "@/lib/realtime";

type GroupDetailProps = {
  groupId: string;
};

type InviteStatus = "none" | "invited" | "pending" | "member";

type InvitableUser = {
  user_id: number;
  first_name: string;
  last_name: string;
  username: string;
  avatar_url: string;
  status: InviteStatus;
};

const statusLabels: Record<Exclude<InviteStatus, "none">, string> = {
  invited: "Invited",
  pending: "Requested",
  member: "Member",
};

type GroupTab = "posts" | "events" | "chat";
type Dialog = "members" | "invite" | "requests" | null;

type ApiResult = { error?: string; message?: string };

async function api<T>(path: string, init?: RequestInit): Promise<{ ok: boolean; status: number; data: T | null }> {
  const response = await fetch(`${API_BASE_URL}/api/groups/${path}`, {
    credentials: "include",
    ...init,
    headers: init?.body ? { "Content-Type": "application/json" } : undefined,
  });
  const data = (await response.json().catch(() => null)) as T | null;
  return { ok: response.ok, status: response.status, data };
}

function displayName(user: { first_name: string; last_name: string; username: string }) {
  return `${user.first_name} ${user.last_name}`.trim() || user.username;
}

function Avatar({ user }: { user: { first_name: string; last_name: string; username: string; avatar_url: string } }) {
  return (
    <span className="group-member-avatar" aria-hidden="true">
      {user.avatar_url ? (
        <img src={imageUrl(user.avatar_url)} alt="" />
      ) : (
        displayName(user).slice(0, 1).toUpperCase()
      )}
    </span>
  );
}

export default function GroupDetail({ groupId }: GroupDetailProps) {
  const id = encodeURIComponent(groupId);
  const [group, setGroup] = useState<Group | null>(null);
  const [loadError, setLoadError] = useState("");
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [requests, setRequests] = useState<GroupMember[]>([]);
  const [invitables, setInvitables] = useState<InvitableUser[]>([]);
  const [dialog, setDialog] = useState<Dialog>(null);
  const [dialogLoading, setDialogLoading] = useState(false);
  const [dialogError, setDialogError] = useState("");
  const [busyId, setBusyId] = useState<number | null>(null);
  const [actionBusy, setActionBusy] = useState(false);
  const [notice, setNotice] = useState("");

  const [activeTab, setActiveTab] = useState<GroupTab>("posts");
  const [eventsCount, setEventsCount] = useState(0);
  const [eventsVersion, setEventsVersion] = useState(0);
  const handleEventsLoaded = useCallback((loadedEvents: GroupEvent[]) => {
    setEventsCount(loadedEvents.length);
  }, []);

  const isAdmin = group?.my_status === "admin";
  const isMember = isAdmin || group?.my_status === "member";

  const loadGroup = useCallback(async () => {
    try {
      const { ok, data } = await api<Group & ApiResult>(id);
      if (!ok || !data) {
        setLoadError(data?.error || "Could not load this group.");
        return;
      }
      setGroup(data);
      setLoadError("");
    } catch {
      setLoadError("Could not load this group.");
    }
  }, [id]);

  const loadMembers = useCallback(async () => {
    try {
      const { ok, data } = await api<GroupMember[]>(`${id}/members`);
      if (ok && Array.isArray(data)) setMembers(data);
    } catch {
      // keep the previous list
    }
  }, [id]);

  const loadRequests = useCallback(async () => {
    try {
      const { ok, data } = await api<GroupMember[]>(`${id}/requests`);
      setRequests(ok && Array.isArray(data) ? data : []);
    } catch {
      // keep the previous list
    }
  }, [id]);

  const loadInvitables = useCallback(async () => {
    try {
      const { ok, data } = await api<InvitableUser[] & ApiResult>(`${id}/invitable`);
      if (!ok) {
        setDialogError((data as ApiResult | null)?.error || "Could not load your followers.");
        setInvitables([]);
        return;
      }
      setInvitables(Array.isArray(data) ? data : []);
    } catch {
      setDialogError("Could not load your followers.");
    }
  }, [id]);

  // first load
  useEffect(() => {
    void loadGroup();
    void loadMembers();
  }, [loadGroup, loadMembers]);

  // only the creator of the group sees pending join requests and can answer them
  useEffect(() => {
    if (isAdmin) void loadRequests();
  }, [isAdmin, loadRequests]);

  // read by the websocket listener without re-subscribing on every change
  const isAdminRef = useRef(isAdmin);
  const dialogRef = useRef(dialog);
  useEffect(() => {
    isAdminRef.current = isAdmin;
    dialogRef.current = dialog;
  }, [isAdmin, dialog]);

  // live updates: someone joined, requested, was invited, was refused...
  useEffect(
    () =>
      subscribe((data) => {
        const reconnected = (data as { type?: string })?.type === "reconnected";
        if (!reconnected && !(isGroupUpdate(data) && String(data.group_id) === String(groupId))) return;
        void loadGroup();
        void loadMembers();
        if (isAdminRef.current) void loadRequests();
        if (dialogRef.current === "invite") void loadInvitables();
      }),
    [groupId, loadGroup, loadMembers, loadRequests, loadInvitables],
  );

  // open a dialog and load what it needs
  async function openDialog(next: Exclude<Dialog, null>) {
    setDialog(next);
    setDialogError("");
    setDialogLoading(true);
    if (next === "members") await loadMembers();
    if (next === "requests") await loadRequests();
    if (next === "invite") await loadInvitables();
    setDialogLoading(false);
  }

  useEffect(() => {
    if (!dialog) return;
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setDialog(null);
    }
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [dialog]);

  async function post(path: string, body?: object) {
    return api<ApiResult>(`${id}/${path}`, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  // join request / invitation answer from the header
  async function headerAction(path: string, body?: object) {
    setActionBusy(true);
    setNotice("");
    try {
      const { ok, data } = await post(path, body);
      setNotice(ok ? data?.message ?? "" : data?.error || "Something went wrong.");
      await Promise.all([loadGroup(), loadMembers()]);
    } catch {
      setNotice("Could not reach the server.");
    } finally {
      setActionBusy(false);
    }
  }

  async function sendInvite(userId: number) {
    setBusyId(userId);
    setDialogError("");
    try {
      const { ok, status, data } = await post("invite", { user_id: userId });
      if (!ok && status !== 409) {
        setDialogError(data?.error || "Could not send the invitation.");
        return;
      }
      setInvitables((users) =>
        users.map((user) => (user.user_id === userId ? { ...user, status: "invited" } : user)),
      );
    } catch {
      setDialogError("Could not send the invitation.");
    } finally {
      setBusyId(null);
    }
  }

  async function answerRequest(userId: number, accept: boolean) {
    setBusyId(userId);
    setDialogError("");
    try {
      const { ok, status, data } = await post(`requests/${userId}`, { accept });
      if (!ok && status !== 404) {
        setDialogError(data?.error || "Could not answer the request.");
        return;
      }
      setRequests((current) => current.filter((r) => r.user_id !== userId));
      if (accept) void Promise.all([loadGroup(), loadMembers()]);
    } catch {
      setDialogError("Could not answer the request.");
    } finally {
      setBusyId(null);
    }
  }

  if (!group) {
    return (
      <main className="group-detail-page">
        <p className="group-invite-empty">{loadError || "Loading…"}</p>
      </main>
    );
  }

  return (
    <main className="group-detail-page">
      <header className="group-detail-header">
        <div className="group-header-content">
          <div className="group-header-copy">
            <p className="group-detail-eyebrow">Community group</p>
            <h1>{group.title}</h1>
            <p className="group-detail-description">{group.description}</p>
            <div className="group-detail-meta">
              {isMember ? (
                <button className="group-meta-link" type="button" onClick={() => void openDialog("members")}>
                  {group.member_count.toLocaleString()} member{group.member_count === 1 ? "" : "s"}
                </button>
              ) : (
                // only members can open the member list
                <span>
                  {group.member_count.toLocaleString()} member{group.member_count === 1 ? "" : "s"}
                </span>
              )}
              <span className="group-meta-divider" aria-hidden="true" />
              <span>Started by {group.admin_name}</span>
            </div>
          </div>

          <div className="group-header-actions">
            {isMember && (
              <button className="group-view-members" type="button" onClick={() => void openDialog("invite")}>
                <span aria-hidden="true">＋</span> Invite people
              </button>
            )}
            {isAdmin && (
              <button
                className="group-secondary-action"
                type="button"
                onClick={() => void openDialog("requests")}
              >
                Join requests
                {requests.length > 0 && <span className="group-count-badge">{requests.length}</span>}
              </button>
            )}
            {group.my_status === "none" && (
              <button
                className="group-view-members"
                type="button"
                disabled={actionBusy}
                onClick={() => void headerAction("join")}
              >
                Request to join
              </button>
            )}
            {group.my_status === "pending" && (
              <button className="group-secondary-action" type="button" disabled>
                Request pending
              </button>
            )}
          </div>
        </div>

        {group.my_status === "invited" && (
          <div className="group-invite-banner" role="status">
            <span>You have been invited to join this group.</span>
            <div>
              <button
                className="group-invite-button"
                type="button"
                disabled={actionBusy}
                onClick={() => void headerAction("invitation", { accept: true })}
              >
                Accept
              </button>
              <button
                className="group-decline-button"
                type="button"
                disabled={actionBusy}
                onClick={() => void headerAction("invitation", { accept: false })}
              >
                Decline
              </button>
            </div>
          </div>
        )}
        {notice && (
          <p className="group-action-notice" role="status">
            {notice}
          </p>
        )}
      </header>

      {isMember ? (
        <section className="group-content" aria-label="Group activity">
          <div className="group-content-heading">
            <div>
              <p className="group-detail-eyebrow">Stay in the loop</p>
              <h2>Community</h2>
            </div>
            <div className="group-tabs" role="tablist" aria-label="Group activity">
              <button
                id="group-posts-tab"
                className={activeTab === "posts" ? "is-active" : ""}
                type="button"
                role="tab"
                aria-selected={activeTab === "posts"}
                aria-controls="group-activity-panel"
                onClick={() => setActiveTab("posts")}
              >
                Posts
              </button>
              <button
                id="group-events-tab"
                className={activeTab === "events" ? "is-active" : ""}
                type="button"
                role="tab"
                aria-selected={activeTab === "events"}
                aria-controls="group-activity-panel"
                onClick={() => setActiveTab("events")}
              >
                Events <span>{eventsCount}</span>
              </button>
              <button
                id="group-chat-tab"
                className={activeTab === "chat" ? "is-active" : ""}
                type="button"
                role="tab"
                aria-selected={activeTab === "chat"}
                aria-controls="group-activity-panel"
                onClick={() => setActiveTab("chat")}
              >
                Chat
              </button>
            </div>
          </div>

          <div
            className="group-activity-panel"
            id="group-activity-panel"
            role="tabpanel"
            aria-labelledby={`group-${activeTab}-tab`}
            tabIndex={0}
          >
            {activeTab === "posts" ? (
              <>
                <CreatePostGp group_id={groupId} />
                <Post query="" myposts group_id={groupId} />
              </>
            ) : activeTab === "chat" ? (
              <GroupChat group_id={groupId} />
            ) : (
              <>
                <CreateEvent group_id={groupId} onEventCreated={() => setEventsVersion((version) => version + 1)} />
                <GetEvents group_id={groupId} refreshKey={eventsVersion} onEventsLoaded={handleEventsLoaded} />
              </>
            )}
          </div>
        </section>
      ) : (
        <section className="group-content group-locked">
          <p>Join this group to see its posts, events and chat.</p>
        </section>
      )}

      {dialog && (
        <div
          className="group-modal-backdrop"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) setDialog(null);
          }}
        >
          <section className="group-members-dialog" role="dialog" aria-modal="true" aria-labelledby="group-dialog-title">
            <div className="group-dialog-heading">
              <div>
                <p className="group-detail-eyebrow">
                  {dialog === "members" ? "The people behind it" : dialog === "invite" ? "Grow the group" : "Waiting for you"}
                </p>
                <h2 id="group-dialog-title">
                  {dialog === "members" ? "Members" : dialog === "invite" ? "Invite people" : "Join requests"}
                </h2>
              </div>
              <button className="group-dialog-close" type="button" aria-label="Close dialog" onClick={() => setDialog(null)}>
                ×
              </button>
            </div>

            <p className="group-dialog-count">
              {dialog === "members"
                ? `${group.member_count.toLocaleString()} member${group.member_count === 1 ? "" : "s"}.`
                : dialog === "invite"
                  ? "Invite your followers and the people you follow."
                  : "As the creator of the group, you can accept or refuse these requests."}
            </p>

            {dialogError && (
              <p className="group-invite-error" role="alert">
                {dialogError}
              </p>
            )}

            {dialogLoading ? (
              <p className="group-invite-empty">Loading…</p>
            ) : dialog === "members" ? (
              <ul className="group-members-list">
                {members.map((member) => (
                  <li key={member.user_id}>
                    <Avatar user={member} />
                    <span className="group-member-identity">
                      <strong>{displayName(member)}</strong>
                      <span>@{member.username}</span>
                    </span>
                    {member.role === "admin" && <span className="group-organizer-badge">Creator</span>}
                  </li>
                ))}
              </ul>
            ) : dialog === "requests" ? (
              requests.length === 0 ? (
                <p className="group-invite-empty">No pending requests.</p>
              ) : (
                <ul className="group-members-list">
                  {requests.map((request) => (
                    <li key={request.user_id}>
                      <Avatar user={request} />
                      <span className="group-member-identity">
                        <strong>{displayName(request)}</strong>
                        <span>@{request.username}</span>
                      </span>
                      <span className="group-row-actions">
                        <button
                          className="group-invite-button"
                          type="button"
                          disabled={busyId === request.user_id}
                          onClick={() => void answerRequest(request.user_id, true)}
                        >
                          Accept
                        </button>
                        <button
                          className="group-decline-button"
                          type="button"
                          disabled={busyId === request.user_id}
                          onClick={() => void answerRequest(request.user_id, false)}
                        >
                          Refuse
                        </button>
                      </span>
                    </li>
                  ))}
                </ul>
              )
            ) : invitables.length === 0 && !dialogError ? (
              <p className="group-invite-empty">No followers or followings to invite yet.</p>
            ) : (
              <ul className="group-members-list">
                {invitables.map((user) => (
                  <li key={user.user_id}>
                    <Avatar user={user} />
                    <span className="group-member-identity">
                      <strong>{displayName(user)}</strong>
                      <span>@{user.username}</span>
                    </span>
                    {user.status === "none" ? (
                      <button
                        className="group-invite-button"
                        type="button"
                        disabled={busyId === user.user_id}
                        onClick={() => void sendInvite(user.user_id)}
                      >
                        {busyId === user.user_id ? "Sending…" : "Invite"}
                      </button>
                    ) : (
                      <span className="group-organizer-badge">{statusLabels[user.status]}</span>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>
      )}
    </main>
  );
}
