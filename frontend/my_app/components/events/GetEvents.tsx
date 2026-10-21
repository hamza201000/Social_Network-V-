"use client";

import { useEffect, useState } from "react";
import type { GroupEvent } from "@/types/eventData";
import EventCard from "./EventCard";
import { API_BASE_URL } from "@/lib/api";

type GetEventsProps = {
  group_id: string;
  refreshKey: number;
  onEventsLoaded?: (events: GroupEvent[]) => void;
};

export default function GetEvents({ group_id, refreshKey, onEventsLoaded }: GetEventsProps) {
  const [events, setEvents] = useState<GroupEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [voteRefreshKey, setVoteRefreshKey] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function loadEvents() {
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(
          `${API_BASE_URL}/api/groups/${encodeURIComponent(group_id)}/events`,
          { credentials: "include" },
        );
        if (!response.ok) throw new Error("Could not load group events.");

        const data: GroupEvent[] = await response.json();
        if (!cancelled) {
          setEvents(data);
          onEventsLoaded?.(data);
        }
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : "Could not load group events.");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void loadEvents();
    return () => {
      cancelled = true;
    };
  }, [group_id, refreshKey, voteRefreshKey, onEventsLoaded]);

  if (loading) return <p className="group-events-status">Loading events…</p>;
  if (error) return <p className="group-events-error" role="alert">{error}</p>;
  if (events.length === 0) {
    return <p className="group-events-status">No events yet. Create the first one for this group.</p>;
  }

  return (
    <div className="group-events-list">
      {events.map((event) => (
        <EventCard
          key={event.id}
          event={event}
          group_id={group_id}
          onVoted={() => setVoteRefreshKey((version) => version + 1)}
        />
      ))}
    </div>
  );
}
