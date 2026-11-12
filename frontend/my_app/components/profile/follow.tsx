"use client"

import { useState } from "react";
import { API_BASE_URL } from "@/lib/api";

export type FollowStatus = "following" | "requested" | "none";

type FollowProps = {
    user_id: number | undefined;
    status: FollowStatus | undefined;
    // called after the status changed, so the profile can reload what is now visible
    onChange?: (status: FollowStatus) => void;
};

const LABELS: Record<FollowStatus, string> = {
    none: "Follow",
    requested: "Requested",
    following: "Unfollow",
};

// Follow a public user, send a follow request to a private user,
// cancel a pending request or unfollow. Give it a `key` with the status so it
// resets when the profile is reloaded.
export default function Follow({ user_id, status, onChange }: FollowProps) {
    const [current, setCurrent] = useState<FollowStatus>(status ?? "none");
    const [confirming, setConfirming] = useState(false);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");

    async function toggleFollow() {
        if (!user_id || busy) return;
        setBusy(true);
        setError("");
        try {
            const response = await fetch(`${API_BASE_URL}/api/user/follow?user_id=${user_id}`, {
                method: "POST",
                credentials: "include",
            });
            const body = await response.json().catch(() => null);
            if (!response.ok) throw new Error(body?.error || "Could not update follow status");
            const next: FollowStatus = body?.data?.follow_status ?? "none";
            setCurrent(next);
            onChange?.(next);
        } catch (followError) {
            setError(followError instanceof Error ? followError.message : "Could not update follow status");
        } finally {
            setBusy(false);
            setConfirming(false);
        }
    }

    if (confirming) {
        return (
            <div className="follow-confirm" role="alertdialog" aria-label="Confirm unfollow">
                <span>Unfollow this user?</span>
                <button className="primary-button" type="button" onClick={toggleFollow} disabled={busy}>
                    Yes, unfollow
                </button>
                <button className="secondary-button" type="button" onClick={() => setConfirming(false)}>
                    Cancel
                </button>
            </div>
        );
    }

    return (
        <div className="follow-action">
            <button
                className={current === "none" ? "primary-button" : "secondary-button"}
                type="button"
                disabled={busy || !user_id}
                title={current === "requested" ? "Click to cancel your follow request" : undefined}
                onClick={() => (current === "following" ? setConfirming(true) : void toggleFollow())}
            >
                {LABELS[current]}
            </button>
            {error && <p className="follow-error" role="alert">{error}</p>}
        </div>
    );
}
