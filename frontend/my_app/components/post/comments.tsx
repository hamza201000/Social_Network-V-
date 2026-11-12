"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import type { CommentData } from "@/types/commentData";
import EmojiPicker from "@/components/chat/EmojiPicker";
import { API_BASE_URL, imageUrl } from "@/lib/api";
import "./comments.css";

const ALLOWED_TYPES = ["image/jpeg", "image/png", "image/gif", "image/webp"];
const MAX_IMAGE_SIZE = 5 * 1024 * 1024;

function commentTime(value: string) {
  return new Date(value).toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

async function fetchComments(postId: number): Promise<CommentData[]> {
  const response = await fetch(`${API_BASE_URL}/api/posts/${postId}/comments`, {
    credentials: "include",
  });
  const data = await response.json().catch(() => null);
  if (!response.ok) throw new Error(data?.error || "Could not load comments.");
  // the backend wraps the list: { message, data: [...] }
  const list = Array.isArray(data) ? data : data?.data;
  return Array.isArray(list) ? list : [];
}

// Comments of one post: the list plus a form to add a comment with an optional image or GIF.
export default function Comments({ postId }: { postId: number }) {
  const [open, setOpen] = useState(false);
  const [comments, setComments] = useState<CommentData[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [content, setContent] = useState("");
  const [image, setImage] = useState<File | null>(null);
  const [preview, setPreview] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const fileInput = useRef<HTMLInputElement>(null);

  async function loadComments() {
    try {
      setComments(await fetchComments(postId));
      setLoaded(true);
      setError("");
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Could not load comments.");
    }
  }

  // load the comments (and so the count) when the post is shown
  useEffect(() => {
    let active = true;
    fetchComments(postId)
      .then((list) => {
        if (!active) return;
        setComments(list);
        setLoaded(true);
      })
      .catch((loadError) => {
        if (active) setError(loadError instanceof Error ? loadError.message : "Could not load comments.");
      });
    return () => {
      active = false;
    };
  }, [postId]);

  useEffect(() => () => {
    if (preview) URL.revokeObjectURL(preview);
  }, [preview]);

  function chooseImage(file: File | undefined) {
    if (!file) return;
    if (!ALLOWED_TYPES.includes(file.type)) {
      setError("Only JPEG, PNG, GIF and WebP images are allowed.");
      return;
    }
    if (file.size > MAX_IMAGE_SIZE) {
      setError("The image must be less than 5 MB.");
      return;
    }
    setError("");
    setImage(file);
    setPreview(URL.createObjectURL(file));
  }

  function removeImage() {
    setImage(null);
    setPreview("");
    if (fileInput.current) fileInput.current.value = "";
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if ((!content.trim() && !image) || sending) return;
    setSending(true);
    setError("");
    const form = new FormData();
    form.append("content", content.trim());
    if (image) form.append("image", image);
    try {
      const response = await fetch(`${API_BASE_URL}/api/posts/${postId}/comments`, {
        method: "POST",
        credentials: "include",
        body: form,
      });
      const data = await response.json().catch(() => null);
      if (!response.ok) throw new Error(data?.error || "Could not add the comment.");
      setContent("");
      removeImage();
      await loadComments();
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : "Could not add the comment.");
    } finally {
      setSending(false);
    }
  }

  return (
    <section className="comments" aria-label="Comments">
      <button
        type="button"
        className="comments-toggle"
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        💬 {loaded ? comments.length : ""} {comments.length === 1 ? "comment" : "comments"}
      </button>

      {open && (
        <>
          {comments.length > 0 && (
            <ul className="comments-list">
              {comments.map((comment) => (
                <li key={comment.id} className="comment">
                  <span className="comment-avatar" aria-hidden="true">
                    {comment.avatar_url ? (
                      <img src={imageUrl(comment.avatar_url)} alt="" />
                    ) : (
                      comment.username.slice(0, 1).toUpperCase()
                    )}
                  </span>
                  <div className="comment-body">
                    <div className="comment-meta">
                      <strong>{comment.username}</strong>
                      <time dateTime={comment.created_at}>{commentTime(comment.created_at)}</time>
                    </div>
                    {comment.content && <p>{comment.content}</p>}
                    {comment.image && (
                      <img
                        className="comment-image"
                        src={imageUrl(comment.image)}
                        alt={`Image in a comment by ${comment.username}`}
                        onError={(event) => {
                          event.currentTarget.hidden = true;
                        }}
                      />
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}

          <form className="comment-form" onSubmit={submit}>
            {preview && (
              <div className="comment-preview">
                <img src={preview} alt="Selected image" />
                <button type="button" onClick={removeImage} aria-label="Remove image">
                  ×
                </button>
              </div>
            )}
            <div className="comment-composer">
              <EmojiPicker onPick={(emoji) => setContent((current) => current + emoji)} />
              <input
                value={content}
                onChange={(event) => setContent(event.target.value)}
                placeholder="Write a comment…"
                aria-label="Write a comment"
                maxLength={2000}
              />
              <input
                ref={fileInput}
                type="file"
                accept="image/jpeg,image/png,image/gif,image/webp"
                hidden
                id={`comment-image-${postId}`}
                onChange={(event) => chooseImage(event.target.files?.[0])}
              />
              <label htmlFor={`comment-image-${postId}`} className="comment-attach" title="Add an image or GIF">
                🖼️
              </label>
              <button type="submit" disabled={sending || (!content.trim() && !image)}>
                {sending ? "…" : "Send"}
              </button>
            </div>
          </form>
        </>
      )}
      {error && <p className="comment-error" role="alert">{error}</p>}
    </section>
  );
}
