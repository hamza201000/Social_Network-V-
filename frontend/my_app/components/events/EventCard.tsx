"use client";

import { useState } from "react";
import type { GroupEvent } from "@/types/eventData";
import type { resData } from "@/types/resData";

type EventCardProps = {
  event: GroupEvent;
  group_id: string;
  onVoted: () => void;
};

export default function EventCard({ event, group_id, onVoted }: EventCardProps) {
  const eventDate = new Date(event.event_at);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function vote(responseChoice: "Going" | "Not going") {
    setSubmitting(true);
    setError(null);
    try {
      const response = await fetch(
        `http://localhost:8080/api/groups/${encodeURIComponent(group_id)}/events/${event.id}/response`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ response: responseChoice }),
        },
      );
      const data: resData = await response.json();
      if (!response.ok) throw new Error(data.err || "Could not save your vote.");
      onVoted();
    } catch (voteError) {
      setError(voteError instanceof Error ? voteError.message : "Could not save your vote.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <article className="group-event">
      <div className="group-event-date sage">
        <span>{eventDate.toLocaleString(undefined, { month: "short" }).toUpperCase()}</span>
        <strong>{eventDate.getDate()}</strong>
      </div>
      <div className="group-event-copy">
        <p className="group-event-type">Group event</p>
        <h3>{event.title}</h3>
        <p>{event.description}</p>
        <time className="group-event-location" dateTime={event.event_at}>
          {eventDate.toLocaleString(undefined, {
            weekday: "long",
            month: "short",
            day: "numeric",
            year: "numeric",
            hour: "numeric",
            minute: "2-digit",
          })}
        </time>
      </div>
      <div className="group-event-attendees">
        <span>{event.going_count} going</span>
        <span>{event.not_going_count} not going</span>
      </div>
      <div className="group-event-vote">
        <button type="button" disabled={submitting} onClick={() => void vote("Going")}>
          {submitting ? "Saving…" : "Going"}
        </button>
        <button type="button" disabled={submitting} onClick={() => void vote("Not going")}>
          Not going
        </button>
      </div>
      {error && <p className="event-vote-error" role="alert">{error}</p>}
    </article>
  );
}
