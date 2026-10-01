"use client"

import { useState } from "react";



export default function Follow({ username }: { username: string }) {
    const [Followed, setFollowed] = useState(false);
    


    const followUser = async () => {
        try {
            const response = await fetch(`http://localhost:8080/api/user/${username}/follow`, {
                method: "GET",
                credentials: "include"
            });
            if (!response.ok) {
                throw new Error("Failed to follow user");
            }
            if (response.ok) {
               Followed ?  setFollowed(false) : setFollowed(true);
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