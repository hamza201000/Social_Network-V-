"use client";

import { useCallback, useEffect, useState } from "react";
import CardGroup from "@/components/group/CardGroup";
import CreateGroup, { type NewGroupDetails } from "@/components/group/createGroup";
import SearchGroup from "@/components/group/searchGroup";
import type { Group, GroupStatus } from "@/types/groupData";
import { API_BASE_URL } from "@/lib/api";
import { isGroupUpdate, subscribe } from "@/lib/realtime";
import "./groups.css";





export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  
  const [notice, setNotice] = useState("");

  const getGroups = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/groups`, {
        method: "GET",
        credentials: "include",
      });
      if (!response.ok) {
        console.error("Failed to fetch group data");
        return;
      }
      const data: Group[] | null = await response.json();
      setGroups(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Failed to fetch group data", err);
    }
  }, []);

  useEffect(() => {
    void getGroups();
  }, [getGroups]);

  // invitations, accepted requests and new members arrive over the websocket
  useEffect(
    () =>
      subscribe((data) => {
        if (isGroupUpdate(data) || (data as { type?: string })?.type === "reconnected") {
          void getGroups();
        }
      }),
    [getGroups],
  );

  function handleStatusChange(groupId: Group["id"], status: GroupStatus, message: string) {
    setGroups((current) =>
      current.map((group) =>
        group.id === groupId
          ? {
              ...group,
              my_status: status,
              member_count:
                group.my_status !== "member" && status === "member"
                  ? group.member_count + 1
                  : group.member_count,
            }
          : group,
      ),
    );
    if (message) setNotice(message);
  }

  return (
    <main className="groups-page">
      <div className="groups-layout">
        <aside className="groups-sidebar">
          <CreateGroup onCreateGroup={getGroups} />
        </aside>

        <section className="groups-browse" aria-labelledby="browse-groups-title">
          <div className="groups-browse-heading">
            <div>
              <p className="group-eyebrow">Made for sharing</p>
              <h2 id="browse-groups-title">Browse groups</h2>
              <p>There’s a place for every interest.</p>
            </div>
            {/* <span className="groups-results-count">{visibleGroups.length} {visibleGroups.length === 1 ? "group" : "groups"}</span> */}
          </div>
          <SearchGroup />
          {notice && <p className="groups-notice" role="status">{notice}</p>}

          {groups?.length > 0 ? (
            <div className="groups-grid">
              {groups.map((group) => (
                <CardGroup key={group.id} group={group} onStatusChange={handleStatusChange} />
              ))}
            </div>
          ) : (
            <div className="groups-empty-state">
              <span aria-hidden="true">⌕</span>
              <h3>No groups found</h3>
              <p> create your group</p>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
