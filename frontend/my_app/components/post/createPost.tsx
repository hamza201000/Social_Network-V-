"use client";
import { resData } from "@/types/resData";
import { UserData } from "@/types/userData";
import { useRef, useState } from "react";
import Connections from "@/components/profile/activeConnections";

type CreatePostProps = {
  onPostCreated: () => void;
};

export default function CreatePost({ onPostCreated }: CreatePostProps) {
    const [content, setContent] = useState("");
    const [privacy, setPrivacy] = useState("public");
    const [image, setImage] = useState<File | null>(null);
    const [imagePreview, setImagePreview] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [openConnections, setOpenConnections] = useState(false);
    const [userId, setUserId] = useState<number>();
    const [selectedFollowerIds, setSelectedFollowerIds] = useState<number[]>([]);
    const fileInputRef = useRef<HTMLInputElement>(null);

    async function handleOpenConnections() {
        if (userId === undefined) {
            try {
                const response = await fetch("http://localhost:8080/api/myprofile", {
                    credentials: "include"
                });
                if (!response.ok) {
                    throw new Error("Could not load your profile.")
                }
                const profile: UserData = await response.json();
                setUserId(profile.user_id);
            } catch (fetchError) {
                console.error("Error fetching profile for post privacy:", fetchError);
                setError("Could not load your followers. Please try again.");
                return;
            }
        }
        setError(null);
        setOpenConnections(true);
    }

    function toggleFollower(userId: number) {
        setSelectedFollowerIds((selectedIds) => (
            selectedIds.includes(userId)
                ? selectedIds.filter((selectedId) => selectedId !== userId)
                : [...selectedIds, userId]
        ));
    }

    function handleImageChange(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0] || null;
        if (file) {
            const allowedTypes = ["image/jpeg", "image/png", "image/gif", "image/webp"];
            if (!allowedTypes.includes(file.type)) {
                setError("Please select a valid image (JPEG, PNG, GIF, or WebP)");
                return;
            }
            if (file.size > 5 * 1024 * 1024) {
                setError("Image must be less than 5MB");
                return;
            }
            setError(null);
            setImage(file);
            setImagePreview(URL.createObjectURL(file));
        }
    }
    function handleRemoveImage() {
        setImage(null);
        setImagePreview(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    }
    async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (privacy === "private" && selectedFollowerIds.length === 0) {
            setError("Choose at least one follower who can see this post.");
            return;
        }

        setLoading(true);
        setError(null);
        const form = new FormData();
        form.append("content", content);
        form.append("privacy", privacy);
        if (privacy === "private") {
            form.append("follower_ids", selectedFollowerIds.join(","));
        }
        if (image) {
            form.append("image", image);
        }
        try {
            const response = await fetch("http://localhost:8080/api/posts", {
                method: "POST",
                body: form,
                credentials: "include"
            });
            const data: resData = await response.json();
            if (!response.ok) {
                throw new Error(data.err || "Could not create post.");
            }

            setContent("");
            setPrivacy("public");
            setImage(null);
            setImagePreview(null);
            setSelectedFollowerIds([]);
            onPostCreated();
        } catch (submitError) {
            console.error("Error creating post:", submitError);
            setError(submitError instanceof Error ? submitError.message : "Could not create post.");
        } finally {
            setLoading(false);
        }
    }

    return (
        <form className="create-post" onSubmit={handleSubmit}>
            <textarea
                placeholder="What's on your mind?"
                value={content}
                onChange={(e) => setContent(e.target.value)}
            />
            <div className="post-privacy-field">
                <label htmlFor="post-privacy">Who can see this post:</label>
                <select
                    id="post-privacy"
                    value={privacy}
                    onChange={(e) => setPrivacy(e.target.value)}
                >
                    <option value="public">Public</option>
                    <option value="almost_private">Almost private</option>
                    <option value="private">Private</option>
                </select>
                <span>
                    {privacy === "public" && "Visible to everyone"}
                    {privacy === "almost_private" && "Visible to your followers"}
                    {privacy === "private" && "Visible to followers you choose"}
                </span>
            </div>
            {privacy === "private" && (
                <div className="post-private-audience">
                    <button
                        type="button"
                        className="post-audience-button"
                        onClick={handleOpenConnections}
                    >
                        Choose followers
                    </button>
                    <span>
                        {selectedFollowerIds.length
                            ? `${selectedFollowerIds.length} follower${selectedFollowerIds.length === 1 ? "" : "s"} selected`
                            : "No followers selected"}
                    </span>
                </div>
            )}
            {error && <p className="post-form-error" role="alert">{error}</p>}
            {openConnections && (
                <div
                    className="post-connections-backdrop"
                    onMouseDown={(event) => {
                        if (event.target === event.currentTarget) setOpenConnections(false);
                    }}
                >
                    <section
                        className="post-connections-dialog"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="post-connections-title"
                        onKeyDown={(event) => {
                            if (event.key === "Escape") setOpenConnections(false);
                        }}
                    >
                        <header className="post-connections-header">
                            <div>
                                <p className="post-connections-eyebrow">Private audience</p>
                                <h2 id="post-connections-title">Choose followers</h2>
                            </div>
                            <button
                                type="button"
                                className="connections-close"
                                aria-label="Close follower list"
                                onClick={() => setOpenConnections(false)}
                            >
                                ×
                            </button>
                        </header>
                        <ul className="connections-list">
                            <Connections
                                user_id={userId}
                                openList="followers"
                                selectedUserIds={selectedFollowerIds}
                                onToggleUser={toggleFollower}
                            />
                        </ul>
                        <footer className="post-connections-footer">
                            <span>{selectedFollowerIds.length} selected</span>
                            <button type="button" className="primary-button" onClick={() => setOpenConnections(false)}>
                                Done
                            </button>
                        </footer>
                    </section>
                </div>
            )}
            <div className="avatar-upload-wrapper">
                <span className="avatar-upload-label">Profile image</span>
                <div className="avatar-upload-row">
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept="image/jpeg,image/png,image/gif,image/webp"
                        onChange={handleImageChange}
                        className="avatar-file-input"
                        id="avatar-upload"
                    />
                    <label htmlFor="avatar-upload" className="avatar-upload-btn">
                        {imagePreview ? "Change image" : "Choose image"}
                    </label>
                    {imagePreview && (
                        <button
                            type="button"
                            className="avatar-remove-btn"
                            onClick={handleRemoveImage}
                        >
                            Remove
                        </button>
                    )}
                </div>
                {imagePreview && (
                    <div className="avatar-preview-wrapper">
                        <img
                            src={imagePreview}
                            alt="Avatar preview"
                            className="avatar-preview-img"
                        />
                    </div>
                )}
            </div>
            <button className="primary-button" disabled={loading}>
                {loading ? "Posting..." : "Post"}
            </button>
        </form>
    );
}