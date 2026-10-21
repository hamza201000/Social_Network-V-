"use client";

import { useEffect, useMemo, useState } from "react";
import CardGroup from "@/components/group/CardGroup";
import CreateGroup, { type NewGroupDetails } from "@/components/group/createGroup";
import SearchGroup from "@/components/group/searchGroup";
import type { Group } from "@/types/groupData";
import { API_BASE_URL } from "@/lib/api";
import "./groups.css";





export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  
  const [memberGroupIds, setMemberGroupIds] = useState<string[]>([]);
  const [requestedGroupIds, setRequestedGroupIds] = useState<string[]>([]);
  const [notice, setNotice] = useState("");

  // const visibleGroups = useMemo(() => {
  //   const search = query.trim().toLocaleLowerCase();
  //   if (!search) return groups;
  //   return groups.filter((group) =>
  //     `${group.title} ${group.description}  ${group.admin_name}`
  //       .toLocaleLowerCase()
  //       .includes(search),
  //   );
  // }, [groups, query]);
  const getGroups = async () => {
    const response = await fetch(`${API_BASE_URL}/api/groups`, {
      method: "GET",
      credentials: "include",
    });
    if (!response.ok) {
      console.error("Failed to fetch group data");
      return;
    }
    const data: Group[] = await response.json();
    console.log("Fetched groups:", data);
    setGroups(data);
};
  useEffect(() => {
  getGroups();
  },[])
  // function handleCreateGroup(details: NewGroupDetails) {
  //   const newGroup: Group = {
  //     ...details,
  //     id: `group-${Date.now()}`,
  //     admin_name: "You",
  //     memberCount: 1,
  //     color: "#c3dfd2",
  //     requests: [],
  //     invitedUsers: [],
  //   };
  //   setGroups((current) => [newGroup, ...current]);
  //   setMemberGroupIds((current) => [...current, newGroup.id]);
  //   setNotice(`“${newGroup.title}” is ready. Invite people to get the conversation started.`);
  // }

  function handleRequestToJoin(groupId: string) {
    setRequestedGroupIds((current) => [...current, groupId]);
    setNotice("Your request has been sent to the group creator.");
  }

  function handleAcceptRequest(groupId: string, requestId: string) {
    setGroups((current) =>
      current.map((group) => {
        if (group.id !== groupId) return group;
        const acceptedRequest = group.requests.find((request) => request.id === requestId);
        return {
          ...group,
          memberCount: group.memberCount + (acceptedRequest ? 1 : 0),
          requests: group.requests.filter((request) => request.id !== requestId),
        };
      }),
    );
    setNotice("Membership request accepted.");
  }

  function handleDeclineRequest(groupId: string, requestId: string) {
    setGroups((current) =>
      current.map((group) => group.id === groupId
        ? { ...group, requests: group.requests.filter((request) => request.id !== requestId) }
        : group),
    );
    setNotice("Membership request declined.");
  }

  function handleInvite(groupId: string, username: string) {
    setGroups((current) =>
      current.map((group) => group.id === groupId && !group.invitedUsers.includes(username)
        ? { ...group, invitedUsers: [...group.invitedUsers, username] }
        : group),
    );
    setNotice(`An invite was sent to @${username}.`);
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
                <CardGroup
                  key={group.id}
                  group={group}
                  isMember={memberGroupIds.includes(group.id)}
                  isRequested={requestedGroupIds.includes(group.id)}
                  onRequestToJoin={handleRequestToJoin}
                  onAcceptRequest={handleAcceptRequest}
                  onDeclineRequest={handleDeclineRequest}
                  onInvite={handleInvite}
                />
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
