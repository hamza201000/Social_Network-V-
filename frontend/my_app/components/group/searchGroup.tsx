"use client";

import { useDebounce } from "@/hooks/debounce";
import { useEffect, useState } from "react";
import SmallCard from "./smallGroupCard";
import { Group } from "@/types/groupData";
import { API_BASE_URL } from "@/lib/api";


export default function SearchGroup() {
  const [query, setQuery] = useState("");
  const [groups, setGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(false)
  const debouncedQuery = useDebounce(query, 1000);
  const isSearching = loading || query.trim() !== debouncedQuery.trim();

  useEffect(() => {
    const q = debouncedQuery.trim();
    if (!q) {
      setGroups([]);
      return;
    }

    const controller = new AbortController();
    const search = async () => {
      try {
        setLoading(true);
        const response = await fetch(
          `${API_BASE_URL}/api/groups/search?q=${encodeURIComponent(q)}`,
          { credentials: "include", signal: controller.signal }
        );
        const data: Group[] | null = response.ok ? await response.json() : [];
        setGroups(Array.isArray(data) ? data : []);
      } catch (err) {
        if ((err as Error).name === "AbortError") return;
        console.error("error searching groups:", err);
        setGroups([]);
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };
    search();
    return () => controller.abort();
  }, [debouncedQuery]);

  return (
    <>
      <div className="group-search-shell">
        <label className="group-search" htmlFor="group-search-input">
          <input
            id="group-search-input"
            type="text"
            placeholder="Search groups by name, topic, or description"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          {query && (
            <button type="button" aria-label="Clear search" onClick={() => setQuery("")}>
              X
            </button>
          )}
        </label>
        {query.trim() && (
          <div className="group-search-results" aria-live="polite">
            {isSearching ? (
              <p className="group-search-message">Searching groups...</p>
            ) : groups.length > 0 ? (
              groups.map((group) => <SmallCard key={group.id} groupData={group} />)
            ) : (
              <p className="group-search-message">No matching groups found.</p>
            )}
          </div>
        )}
      </div>
    </>
  );
}