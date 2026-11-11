"use client";

import { useEffect, useState } from "react";
import type { Group, GroupRequest } from "@/types/groupData";
import { error } from "console";
import Link from "next/link";
import { API_BASE_URL } from "@/lib/api";

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
  onAcceptRequest,
  onDeclineRequest,
  onInvite,
}: CardGroupProps) {
  const [inviteUsername, setInviteUsername] = useState("");
  const [inviteMessage, setInviteMessage] = useState("");
  const isOwner = group.admin_name === "You";

  function handleInvite(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const username = inviteUsername.trim().replace(/^@/, "");
    if (!username) return;
    onInvite(group.id, username);
    setInviteMessage(`Invite sent to @${username}`);
    setInviteUsername("");
  }
  const onRequestToJoin = async (groupId: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/groups/${groupId}/join`, {
        method: "GET",
        credentials: "include",
      });
      if (response.ok) {
        setInviteMessage(`Join to group has been send`);
      }
    } catch (err) {
      console.log("error join group ", err)
    }
  }

  return (
    <Link href={`group/${group.id}`} >
      <article className="group-card">
        <div className="group-card-content">
          <div className="group-card-title-row">
            <h3>{group.title}</h3>
            {isOwner && <span className="group-owner-badge">Creator</span>}
          </div>
          <p className="group-card-description">{group.description}</p>
          <div className="group-card-meta">
            <span className="group-member-avatars" aria-hidden="true">
              <span>✦</span><span>●</span><span>◆</span>
            </span>
            <span>{group.memberCount} members</span>
            <span className="group-meta-dot" aria-hidden="true">·</span>
            <span>by {group.admin_name}</span>
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
            {/* {group.requests.length > 0 && (
            <span className="group-pending-count">
              {group.requests.length} request{group.requests.length === 1 ? "" : "s"}
            </span>
          )} */}
          </div>

         
        </div>
      </article>
    </Link>
  );
}
