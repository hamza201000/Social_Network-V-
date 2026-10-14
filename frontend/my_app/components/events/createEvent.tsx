"use client";

import { useState } from "react";
import type { FormEvent } from "react";
import type { resData } from "@/types/resData";

type CreateEventProps = {
  group_id: string;
  onEventCreated?: () => void;
};

export default function CreateEvent({ group_id, onEventCreated }: CreateEventProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventAt, setEventAt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await fetch(
        `http://localhost:8080/api/groups/${encodeURIComponent(group_id)}/events`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            title: title.trim(),
            description: description.trim(),
            event_at: new Date(eventAt).toISOString(),
          }),
        },
      );
      const data: resData = await response.json();
      if (!response.ok) {
        throw new Error(data.err || "Could not create event.");
      }

      setTitle("");
      setDescription("");
      setEventAt("");
      setSuccess("Event created successfully.");
      onEventCreated?.();
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Could not create event.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form className="create-event-form" onSubmit={handleSubmit}>
      <h3>Create an event</h3>
      <label>
        Event title
        <input
          type="text"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          maxLength={200}
          required
        />
      </label>
      <label>
        Description
        <textarea
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          rows={3}
        />
      </label>
      <label>
        Day and time
        <input
          type="datetime-local"
          value={eventAt}
          onChange={(event) => setEventAt(event.target.value)}
          required
        />
      </label>
      {error && <p className="event-form-error" role="alert">{error}</p>}
      {success && <p className="event-form-success" role="status">{success}</p>}
      <button type="submit" disabled={loading}>
        {loading ? "Creating event…" : "Create event"}
      </button>
    </form>
  );
}
