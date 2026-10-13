import { resData } from "@/types/resData";
import { useRef, useState } from "react";
import "@/components/post/post.css"

type CreatePostProps = {
  onPostCreated?: () => void;
  group_id:string;
};

export default function CreatePostGp({group_id}:CreatePostProps) {
    const [content, setContent] = useState("");
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
            setLoading(true);
            setError(null);
            const form = new FormData();
            form.append("content", content);

            if (image) {
                form.append("image", image);
            }
            try {
                const response = await fetch(`http://localhost:8080/api/groups/${encodeURIComponent(group_id)}/posts`, {
                    method: "POST",
                    body: form,
                    credentials: "include"
                });
                const data: resData = await response.json();
                if (!response.ok) {
                    throw new Error(data.err || "Could not create post.");
                }
    
                setContent("");
                
                setImage(null);
                setImagePreview(null);
                
                // onPostCreated();
            } catch (submitError) {
                console.error("Error creating post:", submitError);
                setError(submitError instanceof Error ? submitError.message : "Could not create post.");
            } finally {
                setLoading(false);
            }
        }


    return (
        <form className="create-post-group" onSubmit={handleSubmit}>
            <textarea
                placeholder="What's on your mind?"
                value={content}
                onChange={(e) => setContent(e.target.value)}
            />


            {error && <p className="post-form-error" role="alert">{error}</p>}

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
    )



}