"use client";

type SearchGroupProps = {
  query: string;
  onQueryChange: (query: string) => void;
};

export default function SearchGroup({ query, onQueryChange }: SearchGroupProps) {
  return (
    <label className="group-search" htmlFor="group-search-input">
      <svg
        aria-hidden="true"
        className="group-search-icon"
        viewBox="0 0 24 24"
        fill="none"
      >
        <circle cx="10.8" cy="10.8" r="6.8" />
        <path d="m16 16 4.5 4.5" />
      </svg>
      <input
        id="group-search-input"
        type="search"
        placeholder="Search groups by name, topic, or description"
        value={query}
        onChange={(event) => onQueryChange(event.target.value)}
      />
      <span className="group-search-shortcut" aria-hidden="true">⌕</span>
    </label>
  );
}