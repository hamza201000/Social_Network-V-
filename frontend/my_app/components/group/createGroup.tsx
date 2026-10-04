"use client";

import { useState } from "react";

export type NewGroupDetails = {
  title: string;
  description: string;
  id?: string;
};

type CreateGroupProps = {
  onCreateGroup: (group: NewGroupDetails) => void;
};

export default function CreateGroup({ onCreateGroup }: CreateGroupProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      const groupDetails: NewGroupDetails = {
        title: title.trim(),
        description: description.trim(),
      };
      const response = await fetch("http://localhost:8080/api/groups", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(groupDetails),
      });
      if (!response.ok) {
        throw new Error("Failed to create group");
      }
      const createdGroup = await response.json();
      groupDetails.id = createdGroup.id;
      onCreateGroup(groupDetails);
      setTitle("");
      setDescription("");
    } catch (err) {
      console.error("Error creating group:", err);
    }

  }

  return (
    <section className="group-create-panel" aria-labelledby="create-group-title">
      <div className="group-section-heading">
        <div>
          <h2 id="create-group-title">Create a group</h2>
          <p className="group-eyebrow"></p>
        </div>
      </div>


      <form className="group-create-form" onSubmit={handleSubmit}>
        <label className="group-field" htmlFor="group-name">
          Group name
          <input
            id="group-name"
            name="title"
            autoComplete="off"
            maxLength={60}
            placeholder="e.g. Weekend hikers"
            required
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />
        </label>
        <label className="group-field group-description-field" htmlFor="group-description">
          Description
          <textarea
            id="group-description"
            name="description"
            maxLength={280}
            placeholder="What brings this group together?"
            required
            rows={3}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
          />
          <span className="group-character-count">{description.length}/280</span>
        </label>
        <button className="group-primary-button group-create-button" type="submit">
          <span aria-hidden="true">＋</span> Create group
        </button>
      </form>
    </section>
  );
}
