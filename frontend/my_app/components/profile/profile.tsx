"use client";

import { useEffect, useState } from "react";
import "./profile.css"
import Post from "../post/post";
import { UserData } from "@/types/userData";
import Follow from "./follow";
import Connections from "./activeConnections";
import { API_BASE_URL, imageUrl } from "@/lib/api";


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



export default  function Profile({ username, myProfile }: { username: string, myProfile: boolean }) {
    const [openList, setOpenList] = useState<"followers" | "following" | null>(null);
    const [profile, setProfile] = useState<UserData>();
    const [avatarFailed, setAvatarFailed] = useState(false);
    const [privacySaving, setPrivacySaving] = useState(false);

    // const fullName = [profile.firstName, profile.lastName].filter(Boolean).join(" ");
    
    const modalTitle = openList === "followers" ? "Followers" : "Following";
    let apiPath = `/api/user/${username}`;
    if (myProfile) {
        apiPath = `/api/myprofile`;
    }
    console.log("is my profile",myProfile);
    
    const fetchProfile = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}${apiPath}`, {
                method: "GET",
                credentials: "include"
            });
            if (!response.ok) {
                throw new Error("Failed to fetch profile data");
            }
            const data: UserData = await response.json();
            setProfile(data);
            setAvatarFailed(false);
            console.log("Profile data fetched successfully:", data);
        } catch (error) {
            console.error("Error fetching profile data:", error);
        }
    }
    const togglePrivacy = async () => {
        if (!profile || privacySaving) return;
        const next = !profile.is_private;
        setPrivacySaving(true);
        try {
            const response = await fetch(`${API_BASE_URL}/api/myprofile/privacy`, {
                method: "PUT",
                credentials: "include",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ private: next }),
            });
            if (!response.ok) {
                throw new Error("Failed to update privacy");
            }
            setProfile({ ...profile, is_private: next });
        } catch (error) {
            console.error("Error updating privacy:", error);
        } finally {
            setPrivacySaving(false);
        }
    };

    const canView = myProfile || profile?.can_view !== false;

    useEffect(() => {
        fetchProfile();
    }, []);
    return (
        <div className="profile-container">
            <section className="profile-card" aria-labelledby="profile-title">
                <header className="profile-header">
                    <div className="profile-avatar" aria-label="Profile photo">
                        {profile?.avatar_url && profile.avatar_url!="" ? (
                            <img
                                src={`http://localhost:8080/api/images/${encodeURIComponent(profile.avatar_url)}`}
                                alt="Profile"
                                onError={() => setAvatarFailed(true)}
                            />
                        ) : <img
                                src={`http://localhost:8080/api/images/icon_user.png`}
                                alt="Profile"
                                onError={() => setAvatarFailed(true)}
                            />}
                    </div>
                    <div className="profile-identity">
                        <p className="profile-eyebrow">Profile</p>
                        <h1 id="profile-title">{profile?.first_name || "User"}</h1>
                        {profile && (
                            <p className="profile-privacy-badge">{profile.is_private ? "Private profile" : "Public profile"}</p>
                        )}
                    </div>
                    {!myProfile ?  <Follow isFollowed={profile?.followed_by_me}  user_id={profile?.user_id} />:<></> }
                    {myProfile && profile && (
                        <button
                            type="button"
                            className="primary-button privacy-toggle"
                            onClick={togglePrivacy}
                            disabled={privacySaving}
                            aria-pressed={!!profile.is_private}
                        >
                            {profile.is_private ? "Switch to public" : "Switch to private"}
                        </button>
                    )}
                    
                </header>

                <div className="profile-connections" aria-label="Connections">
                    <button type="button" className="profile-connection-button" disabled={!canView} onClick={() => setOpenList("followers")}>
                        <strong>{profile?.followers_count}</strong>
                        <span>Followers</span>
                    </button>
                    <button type="button" className="profile-connection-button" disabled={!canView} onClick={() => setOpenList("following")}>
                        <strong>{profile?.following_count}</strong>
                        <span>Following</span>
                    </button>
                </div>

                <div className="profile-divider" />

                {!canView ? (
                    <p className="profile-private-notice">This profile is private. Follow this user to see their information and posts.</p>
                ) : (
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
                )}
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
            {canView && (
                <div className="profile-card" id="profile-posts">
                    <Post query={username} myposts={myProfile} group_id={undefined}/>
                </div>
            )}
        </div>
    );
}