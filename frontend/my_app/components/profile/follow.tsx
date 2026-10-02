"use client"

import { useEffect, useState } from "react";



export default function Follow({ user_id, isFollowed }: { isFollowed: boolean | undefined, user_id: number | undefined }) {
    const [Followed, setFollowed] = useState(isFollowed);

    useEffect(() => {
        setFollowed(isFollowed);
    }, [isFollowed]);
    const followUser = async () => {
        try {
            const response = await fetch(`http://localhost:8080/api/user/follow${user_id ? `?user_id=${user_id}` : ''}`, {
                method: "GET",
                credentials: "include"
            });
            if (!response.ok) {
                throw new Error("Failed to follow user");
            }
            if (response.ok) {
                 setFollowed(pev=>!pev) 
            }
        } catch (error) {
            console.error("Error following user:", error);
        }
    };



    return (
        <>
            <button className="primary-button" onClick={followUser}>{Followed ? "Unfollow" : "Follow"}</button>
        </>
    )
}