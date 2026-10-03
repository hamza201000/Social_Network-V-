"use client";

import { useState } from "react";
import type { Group, GroupRequest } from "@/types/groupData";

type CardGroupProps = {
  group: Group;
  isMember: boolean;
  isRequested: boolean;
  onRequestToJoin: (groupId: string) => void;
  onAcceptRequest: (groupId: string, requestId: string) => void;
  onDeclineRequest: (groupId: string, requestId: string) => void;
  onInvite: (groupId: string, username: string) => void;
};

function RequestRow({
  request,
  onAccept,
  onDecline,
}: {
  request: GroupRequest;
  onAccept: () => void;
  onDecline: () => void;
}) {
  return (
    <li className="group-request-row">
      <span className="group-request-avatar" aria-hidden="true">
        {request.username.slice(0, 1).toUpperCase()}
      </span>
      <span className="group-request-name">@{request.username}</span>
      <button className="group-request-action group-accept-action" onClick={onAccept} type="button">
        Accept
      </button>
      <button className="group-request-action group-decline-action" onClick={onDecline} type="button">
        Decline
      </button>
    </li>
  );
}

export default function CardGroup({
  group,
  isMember,
  isRequested,
  onRequestToJoin,
  onAcceptRequest,
  onDeclineRequest,
  onInvite,
}: CardGroupProps) {
  const [inviteUsername, setInviteUsername] = useState("");
  const [inviteMessage, setInviteMessage] = useState("");
  const isOwner = group.creator === "You";

  function handleInvite(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const username = inviteUsername.trim().replace(/^@/, "");
    if (!username) return;
    onInvite(group.id, username);
    setInviteMessage(`Invite sent to @${username}`);
    setInviteUsername("");
  }

  return (
    <article className="group-card">
      <div className="group-card-cover" style={{ "--group-color": group.color } as React.CSSProperties}>
        <span className="group-card-monogram" aria-hidden="true">{group.name.slice(0, 1).toUpperCase()}</span>
        <span className="group-category">{group.category}</span>
      </div>
      <div className="group-card-content">
        <div className="group-card-title-row">
          <h3>{group.name}</h3>
          {isOwner && <span className="group-owner-badge">Creator</span>}
        </div>
        <p className="group-card-description">{group.description}</p>
        <div className="group-card-meta">
          <span className="group-member-avatars" aria-hidden="true">
            <span>✦</span><span>●</span><span>◆</span>
          </span>
          <span>{group.memberCount.toLocaleString()} members</span>
          <span className="group-meta-dot" aria-hidden="true">·</span>
          <span>by {group.creator}</span>
        </div>

        <div className="group-card-footer">
          {isOwner || isMember ? (
            <span className="group-member-status">
              <span aria-hidden="true">✓</span> {isOwner ? "Your group" : "Member"}
            </span>
          ) : isRequested ? (
            <button className="group-secondary-button" type="button" disabled>
              Request pending
            </button>
          ) : (
            <button
              className="group-primary-button group-join-button"
              onClick={() => onRequestToJoin(group.id)}
              type="button"
            >
              Request to join <span aria-hidden="true">→</span>
            </button>
          )}
          {group.requests.length > 0 && (
            <span className="group-pending-count">
              {group.requests.length} request{group.requests.length === 1 ? "" : "s"}
            </span>
          )}
        </div>

        {isOwner && (
          <div className="group-owner-tools">
            <form className="group-invite-form" onSubmit={handleInvite}>
              <label className="group-field" htmlFor={`invite-${group.id}`}>Invite someone</label>
              <div className="group-invite-controls">
                <span aria-hidden="true">@</span>
                <input
                  id={`invite-${group.id}`}
                  autoComplete="off"
                  placeholder="username"
                  value={inviteUsername}
                  onChange={(event) => setInviteUsername(event.target.value)}
                />
                <button className="group-invite-button" type="submit" disabled={!inviteUsername.trim()}>
                  Invite
                </button>
              </div>
              {inviteMessage && <p className="group-inline-message" role="status">{inviteMessage}</p>}
              {group.invitedUsers.length > 0 && (
                <p className="group-invited-list">Invited: {group.invitedUsers.map((user) => `@${user}`).join(", ")}</p>
              )}
            </form>
            {group.requests.length > 0 && (
              <section className="group-requests" aria-label="Membership requests">
                <h4>Membership requests <span>{group.requests.length}</span></h4>
                <ul>
                  {group.requests.map((request) => (
                    <RequestRow
                      key={request.id}
                      request={request}
                      onAccept={() => onAcceptRequest(group.id, request.id)}
                      onDecline={() => onDeclineRequest(group.id, request.id)}
                    />
                  ))}
                </ul>
              </section>
            )}
          </div>
        )}
      </div>
    </article>
  );
}
