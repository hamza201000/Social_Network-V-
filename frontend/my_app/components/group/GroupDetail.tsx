"use client";

import { useEffect, useState } from "react";
import type { Group } from "@/types/groupData";
import "./group-detail.css";

type GroupDetailProps = {
  groupId: string;
};

type Member = {
  name: string;
  username: string;
  role: string;
  initials: string;
  color: string;
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

const sampleMembers: Member[] = [
  { name: "Maya Chen", username: "mayatreks", role: "Organizer", initials: "MC", color: "#d7e9d8" },
  { name: "Jordan Lee", username: "jordangoesoutside", role: "Member", initials: "JL", color: "#f3dfc9" },
  { name: "Sam Rivera", username: "samonthetrail", role: "Member", initials: "SR", color: "#dce4f3" },
  { name: "Avery Brooks", username: "averyb", role: "Member", initials: "AB", color: "#f2dada" },
  { name: "Noah Patel", username: "noahfindsviews", role: "Member", initials: "NP", color: "#eee4bd" },
];

const sampleEvents = [
  {
    day: "18",
    month: "MAY",
    title: "Coastal loop and coffee",
    details: "Sunday, 8:30 AM · 6.2 mi · Easy pace",
    location: "Lands End Trailhead, San Francisco",
    attendees: 18,
    accent: "sage",
  },
  {
    day: "24",
    month: "MAY",
    title: "Redwood morning hike",
    details: "Saturday, 9:00 AM · 4.5 mi · Moderate",
    location: "Muir Woods, Mill Valley",
    attendees: 12,
    accent: "clay",
  },
];

function isGroup(value: unknown): value is Group {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<Group>;
  return typeof candidate.title === "string" && typeof candidate.description === "string";
}

export default function GroupDetail({ groupId }: GroupDetailProps) {
  const [group, setGroup] = useState(sampleGroup);
  const [activeTab, setActiveTab] = useState<GroupTab>("posts");
  const [membersOpen, setMembersOpen] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadGroup() {
      try {
        const response = await fetch(`http://localhost:8080/api/groups/${encodeURIComponent(groupId)}`, {
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
    if (!membersOpen) return;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setMembersOpen(false);
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [membersOpen]);

  return (
    <main className="group-detail-page">
      <header className="group-detail-header">
        <div className="group-header-content">
          <div className="group-header-copy">
            <p className="group-detail-eyebrow">Community group</p>
            <h1>{group.title}</h1>
            <p className="group-detail-description">{group.description}</p>
            <div className="group-detail-meta">
              <span className="group-member-stack" aria-hidden="true">
                {sampleMembers.slice(0, 3).map((member) => (
                  <span key={member.username} style={{ backgroundColor: member.color }}>
                    {member.initials.slice(0, 1)}
                  </span>
                ))}
              </span>
              <span><strong>{group.memberCount.toLocaleString()}</strong> members</span>
              <span className="group-meta-divider" aria-hidden="true" />
              <span>Started by {group.admin_name}</span>
            </div>
          </div>
          <button className="group-view-members" type="button" onClick={() => setMembersOpen(true)}>
            <span aria-hidden="true">＋</span> View members
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
              Events <span>{sampleEvents.length}</span>
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
            <div className="group-posts-empty">
              <span className="group-empty-symbol" aria-hidden="true">✳</span>
              <h3>The trail starts here</h3>
              <p>Group conversations and updates will show up here.</p>
            </div>
          ) : (
            <div className="group-events-list">
              {sampleEvents.map((event) => (
                <article className="group-event" key={event.title}>
                  <div className={`group-event-date ${event.accent}`}>
                    <span>{event.month}</span>
                    <strong>{event.day}</strong>
                  </div>
                  <div className="group-event-copy">
                    <p className="group-event-type">Upcoming meetup</p>
                    <h3>{event.title}</h3>
                    <p>{event.details}</p>
                    <span className="group-event-location">{event.location}</span>
                  </div>
                  <div className="group-event-attendees">
                    <span aria-hidden="true">◉</span> {event.attendees} going
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </section>

      {membersOpen && (
        <div
          className="group-modal-backdrop"
          role="presentation"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) setMembersOpen(false);
          }}
        >
          <section
            className="group-members-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="group-members-title"
          >
            <div className="group-dialog-heading">
              <div>
                <p className="group-detail-eyebrow">The people behind it</p>
                <h2 id="group-members-title">Members</h2>
              </div>
              <button
                className="group-dialog-close"
                type="button"
                aria-label="Close members dialog"
                onClick={() => setMembersOpen(false)}
              >
                ×
              </button>
            </div>
            <p className="group-dialog-count">{group.memberCount.toLocaleString()} curious people, and counting.</p>
            <ul className="group-members-list">
              {sampleMembers.map((member) => (
                <li key={member.username}>
                  <span className="group-member-avatar" style={{ backgroundColor: member.color }} aria-hidden="true">
                    {member.initials}
                  </span>
                  <span className="group-member-identity">
                    <strong>{member.name}</strong>
                    <span>@{member.username}</span>
                  </span>
                  {member.role === "Organizer" && <span className="group-organizer-badge">Organizer</span>}
                </li>
              ))}
            </ul>
          </section>
        </div>
      )}
    </main>
  );
}