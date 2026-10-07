"use client";

import { useEffect, useState } from "react";
import "./profile.css"
import Post from "../post/post";
import { UserData } from "@/types/userData";
import Follow from "./follow";
import Connections from "./activeConnections";


export type ProfileConnection = {
    id: string;
    username: string;
    avatarUrl?: string;
};

export type ProfileData = {
    firstName?: string;
    lastName?: string;
    dateOfBirth?: string;
    avatarUrl?: string;
    nickname?: string;
    aboutMe?: string;
    followers_count?:number
    following_count?:number
};

// type ProfileProps = {
//     profile?: ProfileData;
//     followers?: ProfileConnectionimport Follow from "./follow";
// [];
//     following?: ProfileConnection[];
// };

export default  function Profile({ username, myProfile }: { username: string, myProfile: boolean }) {
    const [openList, setOpenList] = useState<"followers" | "following" | null>(null);
    const [profile, setProfile] = useState<UserData>();
    const [followers, setFollowers] = useState<ProfileConnection[]>([]);
    const [following, setFollowing] = useState<ProfileConnection[]>([]);

    // const fullName = [profile.firstName, profile.lastName].filter(Boolean).join(" ");
    const activeConnections = openList === "followers" ? followers : following;
    const modalTitle = openList === "followers" ? "Followers" : "Following";
    let apiUrl = `/api/user/${username}`;
    if (myProfile) {
        apiUrl = `/api/myprofile`;
    }
    console.log("is my profile",myProfile);
    
    const fetchProfile = async () => {
        try {
            const response = await fetch(`http://localhost:8080${apiUrl}`, {
                method: "GET",
                credentials: "include"
            });
            if (!response.ok) {
                throw new Error("Failed to fetch profile data");
            }
            const data: UserData = await response.json();
            setProfile(data);
            console.log("Profile data fetched successfully:", data);
        } catch (error) {
            console.error("Error fetching profile data:", error);
        }
    }
    useEffect(() => {
        fetchProfile();
    }, []);
    return (
        <div className="profile-container">
            <section className="profile-card" aria-labelledby="profile-title">
                <header className="profile-header">
                    <div className="profile-avatar" aria-label={profile?.avatarUrl ? "Profile photo" : "No profile photo added"}>
                        {profile?.avatarUrl ? (
                            <img src={profile.avatarUrl} alt="Profile" />
                        ) : (
                            <span className="profile-avatar-placeholder" aria-hidden="true">
                                +
                            </span>
                        )}
                    </div>
                    <div className="profile-identity">
                        <p className="profile-eyebrow">Profile</p>
                        <h1 id="profile-title">{profile?.first_name || "User"}</h1>
                    </div>
                    {!myProfile ?  <Follow isFollowed={profile?.followed_by_me}  user_id={profile?.user_id} />:<></> }
                    
                </header>

                <div className="profile-connections" aria-label="Connections">
                    <button type="button" className="profile-connection-button" onClick={() => setOpenList("followers")}>
                        <strong>{profile?.followers_count}</strong>
                        <span>Followers</span>
                    </button>
                    <button type="button" className="profile-connection-button" onClick={() => setOpenList("following")}>
                        <strong>{profile?.following_count}</strong>
                        <span>Following</span>
                    </button>
                </div>

                <div className="profile-divider" />

                <dl className="profile-details">
                    <div className="profile-detail">
                        <dt>First name</dt>
                        <dd>{profile?.first_name || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Last name</dt>
                        <dd>{profile?.last_name || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Date of birth</dt>
                        <dd>{profile?.date_of_birth || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Nickname</dt>
                        <dd>{profile?.username ? `${profile.username}` : "Not provided"}</dd>
                    </div>
                    <div className="profile-detail profile-about">
                        <dt>About me</dt>
                        <dd>{profile?.about_me || "Your story will appear here."}</dd>
                    </div>
                </dl>
            </section>

            {openList && (
                <div className="connections-backdrop" onMouseDown={(event) => {
                    if (event.target === event.currentTarget) setOpenList(null);
                }}>
                    <section
                        className="connections-dialog"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="connections-title"
                        onKeyDown={(event) => {
                            if (event.key === "Escape") setOpenList(null);
                        }}
                    >
                        <header className="connections-dialog-header">
                            <div>
                                <p className="profile-eyebrow">Your community</p>
                                <h2 id="connections-title">{modalTitle}</h2>
                            </div>
                            <button
                                type="button"
                                className="connections-close"
                                aria-label="Close list"
                                onClick={() => setOpenList(null)}
                            >
                                ×
                            </button>
                        </header>
                        <ul className="connections-list">
                            <Connections user_id={profile?.user_id} openList={openList} />
                        </ul>
                    </section>
                </div>
            )}
            <div className="profile-card" id="profile-posts">
                <Post query={username} myposts={myProfile} />
            </div>
        </div>
    );
}