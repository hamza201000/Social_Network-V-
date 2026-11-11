"use client";

import { useDebounce } from "@/hooks/debounce";
import { API_BASE_URL } from "@/lib/api";
import { UserData } from "@/types/userData";
import { useEffect, useState } from "react";
import UserCard from "./smallProfileCard";
import "./searchProfile.css";

export function SearchUser() {
    const [query, setQuery] = useState("");
    const [users, setUsers] = useState<UserData[]>([]);
    const [loading, setLoading] = useState(false);
    const debouncedQuery = useDebounce(query, 1000);
    const isSearching = loading || query.trim() !== debouncedQuery.trim();

    useEffect(() => {
        const q = debouncedQuery.trim();
        if (!q) {
            setUsers([]);
            return;
        }
        // cancel this request if the query changes before it finishes
        const controller = new AbortController();
        const search = async () => {
            try {
                setLoading(true);
                const response = await fetch(
                    `${API_BASE_URL}/api/user/search?q=${encodeURIComponent(q)}`,
                    { credentials: "include", signal: controller.signal }
                );
                const data: UserData[] | null = response.ok ? await response.json() : [];
                setUsers(Array.isArray(data) ? data : []);
            } catch (err) {
                if ((err as Error).name === "AbortError") return;
                console.error("error searching users:", err);
                setUsers([]);
            } finally {
                if (!controller.signal.aborted) setLoading(false);
            }
        };
        search();
        return () => controller.abort();
    }, [debouncedQuery]);

   

    return (
        <div className="user-search-shell">
            <label className="user-search" htmlFor="user-search-input">
                <input
                    id="user-search-input"
                    type="text"
                    placeholder="Search users by name or username"
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                />
                {query && (
                    <button className="user-search-clear" aria-label="Clear search"
                        type="button"
                        onClick={() => setQuery("")}
                    >
                        X
                    </button>
                )}
            </label>
            {query.trim() && (
                <div className="user-search-results" aria-live="polite">
                    {isSearching ? (
                        <p className="user-search-message">Searching users...</p>
                    ) : users.length > 0 ? (
                        users.map((profile) => (
                            <UserCard key={profile.user_id} userData={profile} deleteHandler={() => {
                                setUsers([]);
                                setQuery("");
                            }} />
                        ))
                    ) : (
                        <p className="user-search-message">No matching users found.</p>
                    )}
                </div>
            )}
        </div>
    );
}
