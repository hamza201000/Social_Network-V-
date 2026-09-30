"use client";
import { resData } from "@/types/resData";
import { useRef, useState } from "react";
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
    const fileInputRef = useRef<HTMLInputElement>(null);

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
            setLoading(true)
            const Form = new FormData()
            Form.append("content", content);
            Form.append("privacy", privacy);
            if (image) {
                Form.append("image", image);
            }
            try {
                setLoading(true)
                const response = await fetch(`http://localhost:8080/api/posts`, {
                    method: "POST",
                    body: Form,
                    credentials: "include"
                })
                const data: resData = await response.json()
                if (!response.ok) {
                    const error = data.err
                    if (error != undefined) {
                        setError(error)
                    }
                }
            } catch (err) {
    
            } finally {
                setLoading(false)
                setContent("")
                setPrivacy("public")
                setImage(null)
            }
            onPostCreated()
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
                Post
            </button>
        </form>
    );
}