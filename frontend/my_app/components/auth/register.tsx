
"use client"

import { resData } from "@/types/resData";
import { FORMERR } from "dns";
import { useRef, useState } from "react";
import "./auth.css"
import Link from "next/link";
export default function RegisterForm() {

    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [username, setUsername] = useState("");
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [avatarFile, setAvatarFile] = useState<File | null>(null);
    const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);
    function handleAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
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
            setAvatarFile(file);
            setAvatarPreview(URL.createObjectURL(file));
        }
    }
    function handleRemoveAvatar() {
        setAvatarFile(null);
        setAvatarPreview(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    }
    async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setLoading(true)
        const Form = new FormData()
        Form.append("first_name", firstName);
        Form.append("last_name", lastName);
        Form.append("username", username);
        Form.append("email", email);
        Form.append("password", password);
        if (avatarFile) {
            Form.append("avatar", avatarFile);
            
        } 
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/register`, {
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
        }

    }

    return (
        <section className="auth-card" aria-label="Authentication form">
            <div className="auth-card-header">
                <p className="eyebrow">
                    Start fresh
                </p>
                <h1>
                    Create your account
                </h1>
                <p>


                    Join the space and start building your profile.
                </p>
            </div>

            <form className="auth-form" onSubmit={handleSubmit}>
                <div className="auth-name-grid">
                    <label>
                        First name
                        <input
                            value={firstName}
                            onChange={(e) => setFirstName(e.target.value)}
                            placeholder="Mina"
                        />
                    </label>
                    <label>
                        Last name
                        <input
                            value={lastName}
                            onChange={(e) => setLastName(e.target.value)}
                            placeholder="Stone"
                        />
                    </label>
                    <label>
                        Username
                        <input
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="mina"
                        />
                    </label>
                    {/* Avatar upload */}
                    <div className="avatar-upload-wrapper">
                        <span className="avatar-upload-label">Profile image</span>
                        <div className="avatar-upload-row">
                            <input
                                ref={fileInputRef}
                                type="file"
                                accept="image/jpeg,image/png,image/gif,image/webp"
                                onChange={handleAvatarChange}
                                className="avatar-file-input"
                                id="avatar-upload"
                            />
                            <label htmlFor="avatar-upload" className="avatar-upload-btn">
                                {avatarPreview ? "Change image" : "Choose image"}
                            </label>
                            {avatarPreview && (
                                <button
                                    type="button"
                                    className="avatar-remove-btn"
                                    onClick={handleRemoveAvatar}
                                >
                                    Remove
                                </button>
                            )}
                        </div>
                        {avatarPreview && (
                            <div className="avatar-preview-wrapper">
                                {/* eslint-disable-next-line @next/next/no-img-element */}
                                <img
                                    src={avatarPreview}
                                    alt="Avatar preview"
                                    className="avatar-preview-img"
                                />
                            </div>
                        )}
                    </div>
                    <label>
                        Email
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="mina@example.com"
                        />
                    </label>
                    <label>
                        Password
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="Password"
                        />
                    </label>
                </div>
                {error && <div className="error">{String(error)}</div>}
                <button className="primary-button" disabled={loading}>
                    Create account
                </button>
                <p className="auth-footer">
                    Already have an account?
                    <Link href={"/login"}>
                        Sign in
                    </Link>
                </p>
            </form>
        </section >
    )


}