"use client";

import { useState } from "react";

export type NewGroupDetails = {
  name: string;
  description: string;
  category: string;
};

type CreateGroupProps = {
  onCreateGroup: (group: NewGroupDetails) => void;
};

export default function CreateGroup({ onCreateGroup }: CreateGroupProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState("Community");

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onCreateGroup({ name: name.trim(), description: description.trim(), category });
    setName("");
    setDescription("");
    setCategory("Community");
  }

  return (
    <section className="group-create-panel" aria-labelledby="create-group-title">
      <div className="group-section-heading">
        <span className="group-heading-icon" aria-hidden="true">+</span>
        <div>
          <p className="group-eyebrow">Make it yours</p>
          <h2 id="create-group-title">Create a group</h2>
        </div>
      </div>
      <p className="group-panel-description">
        Start a space for people who love the same things you do.
      </p>

      <form className="group-create-form" onSubmit={handleSubmit}>
        <label className="group-field" htmlFor="group-name">
          Group name
          <input
            id="group-name"
            name="name"
            autoComplete="off"
            maxLength={60}
            placeholder="e.g. Weekend hikers"
            required
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </label>
        <label className="group-field" htmlFor="group-category">
          Category
          <select
            id="group-category"
            value={category}
            onChange={(event) => setCategory(event.target.value)}
          >
            <option>Community</option>
            <option>Arts &amp; culture</option>
            <option>Books</option>
            <option>Food</option>
            <option>Outdoors</option>
            <option>Technology</option>
          </select>
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
