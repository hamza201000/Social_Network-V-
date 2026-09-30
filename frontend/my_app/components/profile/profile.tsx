"use client";

import { useState } from "react";
import "./profile.css"
import Post from "../post/post";

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
};

type ProfileProps = {
    profile?: ProfileData;
    followers?: ProfileConnection[];
    following?: ProfileConnection[];
};

export default function Profile({ profile = {}, followers = [], following = [] }: ProfileProps) {
    const [openList, setOpenList] = useState<"followers" | "following" | null>(null);
    const fullName = [profile.firstName, profile.lastName].filter(Boolean).join(" ");
    const activeConnections = openList === "followers" ? followers : following;
    const modalTitle = openList === "followers" ? "Followers" : "Following";

    return (
        <div className="profile-container">
            <section className="profile-card" aria-labelledby="profile-title">
                <header className="profile-header">
                    <div className="profile-avatar" aria-label={profile.avatarUrl ? "Profile photo" : "No profile photo added"}>
                        {profile.avatarUrl ? (
                            <img src={profile.avatarUrl} alt="Profile" />
                        ) : (
                            <span className="profile-avatar-placeholder" aria-hidden="true">
                                {fullName ? fullName.slice(0, 1).toUpperCase() : "+"}
                            </span>
                        )}
                    </div>
                    <div className="profile-identity">
                        <p className="profile-eyebrow">Profile</p>
                        <h1 id="profile-title">{fullName}</h1>
                        <p className="profile-nickname">
                            {profile.nickname ? `@${profile.nickname.replace(/^@/, "")}` : "Nickname not added"}
                        </p>
                    </div>
                </header>

                <div className="profile-connections" aria-label="Connections">
                    <button type="button" className="profile-connection-button" onClick={() => setOpenList("followers")}>
                        <strong>{followers.length}</strong>
                        <span>Followers</span>
                    </button>
                    <button type="button" className="profile-connection-button" onClick={() => setOpenList("following")}>
                        <strong>{following.length}</strong>
                        <span>Following</span>
                    </button>
                </div>

                <div className="profile-divider" />

                <dl className="profile-details">
                    <div className="profile-detail">
                        <dt>First name</dt>
                        <dd>{profile.firstName || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Last name</dt>
                        <dd>{profile.lastName || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Date of birth</dt>
                        <dd>{profile.dateOfBirth || "Not provided"}</dd>
                    </div>
                    <div className="profile-detail">
                        <dt>Nickname</dt>
                        <dd>{profile.nickname ? `@${profile.nickname.replace(/^@/, "")}` : "Not provided"}</dd>
                    </div>
                    <div className="profile-detail profile-about">
                        <dt>About me</dt>
                        <dd>{profile.aboutMe || "Your story will appear here."}</dd>
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
                            {activeConnections.length ? activeConnections.map((connection) => (
                                <li key={connection.id} className="connection-item">
                                    <div className="connection-avatar" aria-hidden="true">
                                        {connection.avatarUrl ? (
                                            
                                            <img src={connection.avatarUrl} alt="" />
                                        ) : (
                                            <span>{connection.username.slice(0, 1).toUpperCase()}</span>
                                        )}
                                    </div>
                                    <span className="connection-username">@{connection.username.replace(/^@/, "")}</span>
                                </li>
                            )) : (
                                <li className="connections-empty">No {openList} to show yet.</li>
                            )}
                        </ul>
                    </section>
                </div>
            )}
            <div className="profile-card" id="profile-posts"> 
                <Post query={"goku"} />
            </div>
        </div>
    );
}