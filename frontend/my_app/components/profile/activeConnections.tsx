import { ProfileConnection } from "@/types/userData"
import { useEffect, useState } from "react"
import ConnectedProfiles from "./ConnectedProfiles"
import { API_BASE_URL } from "@/lib/api";

export default function Connections({
    user_id,
    openList,
    selectedUserIds,
    onToggleUser,
}: {
    user_id: number | undefined;
    openList: string;
    selectedUserIds?: number[];
    onToggleUser?: (userId: number) => void;
}) {
    const [followers, setFollowers] = useState<ProfileConnection[]>([])
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (user_id === undefined) return

        const fetchFollowers = async () => {
            try {
                setError(null)
                const response = await fetch(`${API_BASE_URL}/api/user/${user_id}/${openList}`, {
                    credentials: "include"
                })
                if (!response.ok) {
                    throw new Error("Could not load connections.")
                }
                const data: ProfileConnection[] = await response.json()
                setFollowers(data)
            } catch (fetchError) {
                console.error("Error fetching connections:", fetchError)
                setError("Could not load connections. Please try again.")
            }
        }
        fetchFollowers()
    }, [user_id, openList])

    return (
        <>
            {error ? (
                <li className="connections-empty" role="alert">{error}</li>
            ) : followers?.length ? followers.map((connection) => (
                <ConnectedProfiles
                    key={connection.user_id}
                    connection={connection}
                    isSelected={selectedUserIds?.includes(connection.user_id)}
                    onSelect={onToggleUser}
                />
            )) : (
                <li className="connections-empty">No followers to show yet.</li>
            )}
        </>
    )
}