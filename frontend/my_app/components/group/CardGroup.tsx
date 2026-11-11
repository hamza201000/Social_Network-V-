"use client";

import { useState } from "react";
import Link from "next/link";
import type { Group, GroupStatus } from "@/types/groupData";
import { API_BASE_URL } from "@/lib/api";

type CardGroupProps = {
  group: Group;
  onStatusChange: (groupId: Group["id"], status: GroupStatus, notice: string) => void;
};

export default function CardGroup({ group, onStatusChange }: CardGroupProps) {
  const [busy, setBusy] = useState(false);
  const isOwner = group.my_status === "admin";

  async function call(path: string, body?: object) {
    setBusy(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/groups/${group.id}/${path}`, {
        method: "POST",
        credentials: "include",
        headers: body ? { "Content-Type": "application/json" } : undefined,
        body: body ? JSON.stringify(body) : undefined,
      });
      const data = (await response.json().catch(() => null)) as
        | { error?: string; message?: string; data?: { status?: GroupStatus } }
        | null;
      if (!response.ok) {
        onStatusChange(group.id, group.my_status, data?.error || "Something went wrong.");
        return;
      }
      onStatusChange(group.id, data?.data?.status ?? group.my_status, data?.message ?? "");
    } catch {
      onStatusChange(group.id, group.my_status, "Could not reach the server.");
    } finally {
      setBusy(false);
    }
  }

  // the card is a link: buttons inside it must not navigate
  function action(event: React.MouseEvent, run: () => void) {
    event.preventDefault();
    event.stopPropagation();
    run();
  }

  return (
    <Link href={`group/${group.id}`}>
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
            <span>
              {group.member_count} member{group.member_count === 1 ? "" : "s"}
            </span>
            <span className="group-meta-dot" aria-hidden="true">·</span>
            <span>by {group.admin_name}</span>
          </div>

          <div className="group-card-footer">
            {group.my_status === "admin" || group.my_status === "member" ? (
              <span className="group-member-status">
                <span aria-hidden="true">✓</span> {isOwner ? "Your group" : "Member"}
              </span>
            ) : group.my_status === "pending" ? (
              <button className="group-secondary-button" type="button" disabled>
                Request pending
              </button>
            ) : group.my_status === "invited" ? (
              <>
                <button
                  className="group-primary-button group-join-button"
                  type="button"
                  disabled={busy}
                  onClick={(e) => action(e, () => void call("invitation", { accept: true }))}
                >
                  Accept invite
                </button>
                <button
                  className="group-secondary-button"
                  type="button"
                  disabled={busy}
                  onClick={(e) => action(e, () => void call("invitation", { accept: false }))}
                >
                  Decline
                </button>
              </>
            ) : (
              <button
                className="group-primary-button group-join-button"
                type="button"
                disabled={busy}
                onClick={(e) => action(e, () => void call("join"))}
              >
                Request to join <span aria-hidden="true">→</span>
              </button>
            )}
          </div>
        </div>
      </article>
    </Link>
  );
}
