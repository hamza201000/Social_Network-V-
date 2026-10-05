"use client";

import { useDebounce } from "@/hooks/debounce";
import { useEffect, useState } from "react";
import SmallCard from "./smallGroupCard";
import { Group } from "@/types/groupData";


type SearchGroupProps = {
  query: string;
  onQueryChange: (query: string) => void;
};

export default function SearchGroup() {
  const [query, setQuery] = useState("");
  const [groups, setGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(false)
  const debouncedQuery = useDebounce(query, 1000);
  const search = async (q: string) => {
    try {
      setLoading(true)
      const response = await fetch(
        `http://localhost:8080/api/groups/search?q=${encodeURIComponent(q)}`, {
        credentials: "include",
      }
      );
      const data: Group[] = await response.json();
      console.log("data shearch", data);
      setGroups(data)
    } catch (err) {
      setLoading(false)
      throw console.error("error shearch data");
    } finally {
      setLoading(false)
    }
  }
  useEffect(() => {
    if (debouncedQuery.trim()) search(debouncedQuery)
  }, [debouncedQuery])


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
          {query&&<button 
            onClick={()=>{
              setQuery("")
              setGroups([])
            }}
          
          >X</button>}
        </label>
        {query.trim() && (
          <div className="group-search-results" aria-live="polite">
            {loading ? (
              <p className="group-search-message">Searching groups...</p>
            ) : groups?.length > 0 ? (
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