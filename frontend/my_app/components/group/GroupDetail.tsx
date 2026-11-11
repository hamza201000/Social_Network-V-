"use client";

import { useCallback, useEffect, useState } from "react";
import type { Group } from "@/types/groupData";
import type { GroupEvent } from "@/types/eventData";
import "./group-detail.css";
import CreatePostGp from "../post/createPostGp";
import Post from "../post/post";
import CreateEvent from "@/components/events/createEvent";
import GetEvents from "@/components/events/GetEvents";
import { API_BASE_URL } from "@/lib/api";

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

type GroupTab = "posts" | "events";

const sampleGroup: Group = {
  id: "weekend-trail-club",
  title: "Weekend Trail Club",
  description:
    "Fresh air, good company, and a new trail every weekend. All paces and experience levels are welcome.",
  admin_name: "Maya Chen",
  memberCount: 1240,
  color: "#6f9f78",
  requests: [],
  invitedUsers: [],
};

function isGroup(value: unknown): value is Group {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<Group>;
  return typeof candidate.title === "string" && typeof candidate.description === "string";
}

export default function GroupDetail({ groupId }: GroupDetailProps) {
  const [group, setGroup] = useState(sampleGroup);
  const [activeTab, setActiveTab] = useState<GroupTab>("posts");
  const [inviteOpen, setInviteOpen] = useState(false);
  const [invitables, setInvitables] = useState<InvitableUser[]>([]);
  const [inviteLoading, setInviteLoading] = useState(false);
  const [inviteError, setInviteError] = useState("");
  const [sendingTo, setSendingTo] = useState<number | null>(null);
  const [eventsCount, setEventsCount] = useState(0);
  const [eventsVersion, setEventsVersion] = useState(0);
  const handleEventsLoaded = useCallback((loadedEvents: GroupEvent[]) => {
    setEventsCount(loadedEvents.length);
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function loadGroup() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/groups/${encodeURIComponent(groupId)}`, {
          credentials: "include",
        });
        if (!response.ok) return;

        const data: unknown = await response.json();
        const result = Array.isArray(data) ? data[0] : data;
        console.log("data group ::",data);
        
        if (!cancelled && isGroup(result)) {
          setGroup({ ...sampleGroup, ...result, id: result.id || groupId });
        }
      } catch {
        // Keep the sample group visible when the API is unavailable.
      }
    }

    void loadGroup();
    return () => {
      cancelled = true;
    };
  }, [groupId]);

  useEffect(() => {
    if (!inviteOpen) return;
    let cancelled = false;

    async function loadInvitables() {
      setInviteLoading(true);
      setInviteError("");
      try {
        const response = await fetch(
          `${API_BASE_URL}/api/groups/${encodeURIComponent(groupId)}/invitable`,
          { credentials: "include" },
        );
        const data: unknown = await response.json().catch(() => null);
        if (cancelled) return;
        if (!response.ok) {
          const message = (data as { error?: string } | null)?.error;
          setInviteError(message || "Could not load your followers.");
          setInvitables([]);
          return;
        }
        setInvitables(Array.isArray(data) ? (data as InvitableUser[]) : []);
      } catch {
        if (!cancelled) setInviteError("Could not load your followers.");
      } finally {
        if (!cancelled) setInviteLoading(false);
      }
    }

    void loadInvitables();
    return () => {
      cancelled = true;
    };
  }, [inviteOpen, groupId]);

  async function sendInvite(userId: number) {
    setSendingTo(userId);
    setInviteError("");
    try {
      const response = await fetch(`${API_BASE_URL}/api/groups/${encodeURIComponent(groupId)}/invite`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: userId }),
      });
      if (!response.ok && response.status !== 409) {
        const data = (await response.json().catch(() => null)) as { error?: string } | null;
        setInviteError(data?.error || "Could not send the invitation.");
        return;
      }
      setInvitables((users) =>
        users.map((user) => (user.user_id === userId ? { ...user, status: "invited" } : user)),
      );
    } catch {
      setInviteError("Could not send the invitation.");
    } finally {
      setSendingTo(null);
    }
  }

  useEffect(() => {
    if (!inviteOpen) return;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setInviteOpen(false);
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [inviteOpen]);

  return (
    <main className="group-detail-page">
      <header className="group-detail-header">
        <div className="group-header-content">
          <div className="group-header-copy">
            <p className="group-detail-eyebrow">Community group</p>
            <h1>{group.title}</h1>
            <p className="group-detail-description">{group.description}</p>
            <div className="group-detail-meta">
              <span>
                
                  {group.memberCount.toLocaleString()}
                   members</span>
              <span className="group-meta-divider" aria-hidden="true" />
              <span>Started by {group.admin_name}</span>
            </div>
          </div>
          <button className="group-view-members" type="button" onClick={() => setInviteOpen(true)}>
            <span aria-hidden="true">＋</span> Invite people
          </button>
        </div>
      </header>

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
          </div>
        </div>

        <div
          className="group-activity-panel"
          id="group-activity-panel"
          role="tabpanel"
          aria-labelledby={activeTab === "posts" ? "group-posts-tab" : "group-events-tab"}
          tabIndex={0}
        >
          {activeTab === "posts" ? (
            <>
            <CreatePostGp group_id={groupId}  />
            {/* <div className="group-posts-empty"> */}
              <Post query="" myposts group_id={groupId}/>
            {/* </div> */}
            </>
          ) : (
            <>
              <CreateEvent group_id={groupId} onEventCreated={() => setEventsVersion((version) => version + 1)} />
              <GetEvents group_id={groupId} refreshKey={eventsVersion} onEventsLoaded={handleEventsLoaded} />
            </>
          )}
        </div>
      </section>

      {inviteOpen && (
        <div
          className="group-modal-backdrop"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) setInviteOpen(false);
          }}
        >
          <section
            className="group-members-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="group-invite-title"
          >
            <div className="group-dialog-heading">
              <div>
                <p className="group-detail-eyebrow">Grow the group</p>
                <h2 id="group-invite-title">Invite people</h2>
              </div>
              <button
                className="group-dialog-close"
                type="button"
                aria-label="Close invite dialog"
                onClick={() => setInviteOpen(false)}
              >
                ×
              </button>
            </div>
            <p className="group-dialog-count">Invite your followers and the people you follow.</p>
            {inviteError && (
              <p className="group-invite-error" role="alert">
                {inviteError}
              </p>
            )}
            {inviteLoading ? (
              <p className="group-invite-empty">Loading…</p>
            ) : invitables.length === 0 && !inviteError ? (
              <p className="group-invite-empty">No followers or followings to invite yet.</p>
            ) : (
              <ul className="group-members-list">
                {invitables.map((user) => {
                  const fullName = `${user.first_name} ${user.last_name}`.trim() || user.username;
                  return (
                    <li key={user.user_id}>
                      <span className="group-member-avatar" aria-hidden="true">
                        {user.avatar_url ? (
                          <img
                            src={`${API_BASE_URL}/api/images/${encodeURIComponent(user.avatar_url)}`}
                            alt=""
                          />
                        ) : (
                          fullName.slice(0, 1).toUpperCase()
                        )}
                      </span>
                      <span className="group-member-identity">
                        <strong>{fullName}</strong>
                        <span>@{user.username}</span>
                      </span>
                      {user.status === "none" ? (
                        <button
                          className="group-invite-button"
                          type="button"
                          disabled={sendingTo === user.user_id}
                          onClick={() => void sendInvite(user.user_id)}
                        >
                          {sendingTo === user.user_id ? "Sending…" : "Invite"}
                        </button>
                      ) : (
                        <span className="group-organizer-badge">{statusLabels[user.status]}</span>
                      )}
                    </li>
                  );
                })}
              </ul>
            )}
          </section>
        </div>
      )}
    </main>
  );
}
