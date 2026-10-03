"use client";

import { useMemo, useState } from "react";
import CardGroup from "@/components/group/CardGroup";
import CreateGroup, { type NewGroupDetails } from "@/components/group/createGroup";
import SearchGroup from "@/components/group/searchGroup";
import type { Group } from "@/types/groupData";
import "./groups.css";

const initialGroups: Group[] = [
  {
    id: "weekend-trail-club",
    name: "Weekend Trail Club",
    description: "Fresh air, good company, and a new trail every weekend. All paces and experience levels are welcome.",
    category: "Outdoors",
    creator: "Maya Chen",
    memberCount: 1240,
    color: "#b7d9bf",
    requests: [],
    invitedUsers: [],
  },
  {
    id: "little-book-nook",
    name: "The Little Book Nook",
    description: "A cozy corner of the internet for our monthly reads, thoughtful recommendations, and friendly chats.",
    category: "Books",
    creator: "Jordan Lee",
    memberCount: 864,
    color: "#e5c6a8",
    requests: [],
    invitedUsers: [],
  },
  {
    id: "makers-table",
    name: "The Makers’ Table",
    description: "Share what you’re making, swap ideas, and find inspiration from creative people around the world.",
    category: "Arts & culture",
    creator: "Sam Rivera",
    memberCount: 532,
    color: "#d6c4e7",
    requests: [],
    invitedUsers: [],
  },
];

export default function GroupPage() {
  const [groups, setGroups] = useState(initialGroups);
  const [query, setQuery] = useState("");
  const [memberGroupIds, setMemberGroupIds] = useState<string[]>([]);
  const [requestedGroupIds, setRequestedGroupIds] = useState<string[]>([]);
  const [notice, setNotice] = useState("");

  const visibleGroups = useMemo(() => {
    const search = query.trim().toLocaleLowerCase();
    if (!search) return groups;
    return groups.filter((group) =>
      `${group.name} ${group.description} ${group.category} ${group.creator}`
        .toLocaleLowerCase()
        .includes(search),
    );
  }, [groups, query]);

  function handleCreateGroup(details: NewGroupDetails) {
    const newGroup: Group = {
      ...details,
      id: `group-${Date.now()}`,
      creator: "You",
      memberCount: 1,
      color: "#c3dfd2",
      requests: [],
      invitedUsers: [],
    };
    setGroups((current) => [newGroup, ...current]);
    setMemberGroupIds((current) => [...current, newGroup.id]);
    setNotice(`“${newGroup.name}” is ready. Invite people to get the conversation started.`);
  }

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
      <header className="groups-hero">
        <div className="groups-hero-copy">
          <p className="group-eyebrow">Find your people</p>
          <h1>Groups make good things better.</h1>
          <p>Discover a community, share what you love, and make something together.</p>
        </div>
        <div className="groups-hero-art" aria-hidden="true">
          <span>✳</span><span>✦</span><span>●</span><span>✺</span>
        </div>
      </header>

      <div className="groups-layout">
        <aside className="groups-sidebar">
          <CreateGroup onCreateGroup={handleCreateGroup} />
        </aside>

        <section className="groups-browse" aria-labelledby="browse-groups-title">
          <div className="groups-browse-heading">
            <div>
              <p className="group-eyebrow">Made for sharing</p>
              <h2 id="browse-groups-title">Browse groups</h2>
              <p>There’s a place for every interest.</p>
            </div>
            <span className="groups-results-count">{visibleGroups.length} {visibleGroups.length === 1 ? "group" : "groups"}</span>
          </div>
          <SearchGroup query={query} onQueryChange={setQuery} />
          {notice && <p className="groups-notice" role="status">{notice}</p>}

          {visibleGroups.length > 0 ? (
            <div className="groups-grid">
              {visibleGroups.map((group) => (
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
              <p>Try a different search, or create a group of your own.</p>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
