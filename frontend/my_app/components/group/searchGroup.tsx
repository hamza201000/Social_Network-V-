"use client";

import { useDebounce } from "@/hooks/debounce";
import { error } from "console";
import { useEffect, useState } from "react";


// type SearchGroupProps = {
//   query: string;
//   onQueryChange: (query: string) => void;
// };

export default function SearchGroup() {
  const [query, setQeury] = useState("")

  const debouncedQuery = useDebounce(query, 300);
  const search = async (q:string) => {
    try {

      const response = await fetch(
        `http://localhost:8080/api/groups/search?q=${encodeURIComponent(q)}`, {
        credentials: "include",
      }
      );
      const data = await response.json();
      console.log("data shearch",data);
    } catch (err) {
      throw console.error("error shearch data");
    }
  }
  useEffect(() => {
    if (debouncedQuery.trim()) search(debouncedQuery)
    }, [debouncedQuery])


  return (
    <label className="group-search" htmlFor="group-search-input">
      <input
        id="group-search-input"
        type="search"
        placeholder="Search groups by name, topic, or description"
        value={query}
        onChange={(event) => setQeury(event.target.value)}
      />
    </label>
  );
}