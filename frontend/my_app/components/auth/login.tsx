"use client"
import { resData } from "@/types/resData";
import { useState } from "react";
import "./auth.css"
import Link from "next/link";
import { useRouter } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";


export default function LoginForm() {
    const router=useRouter()
    const [username, setUsername] = useState("")
    const [password, setPassword] = useState("")
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setError(null)
        const login = username.trim()
        if (!login || !password) {
            setError("Please enter your email or username and your password.")
            return
        }
        setLoading(true)
        try {
            const response = await fetch(`${API_BASE_URL}/api/login`, {
                method: "POST",
                headers: { "content-type": "application/json" },
                body: JSON.stringify({ email: login, password }),
                credentials: "include"
            })
            if (!response.ok) {
                const data: resData | null = await response.json().catch(() => null)
                if (response.status === 401) {
                    setError("Wrong email/username or password.")
                } else {
                    setError(data?.error || data?.err || "Something went wrong, please try again.")
                }
                return
            }
            router.replace("/")
            router.refresh()
        } catch {
            setError("Cannot reach the server. Check your connection and try again.")
        } finally {
            setLoading(false)
        }
    }

    return (
        <section className="auth-card" aria-label="Authentication form">
            <div className="auth-card-header">
                <p className="eyebrow">
                    Welcome back
                </p>
                <h1>
                    Sign in to Network
                </h1>
                <p>

                    Continue to your feed, messages, and saved conversations.

                </p>
            </div>
            <form className="auth-form" onSubmit={handleSubmit}>
                <label>
                    Email or Username
                    <input
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        placeholder="goku or goku@example.com"
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
                {error && <div className="error">{String(error)}</div>}
                <button className="primary-button" disabled={loading}>
                    {loading ? "Signing in..." : "Sign in"}
                </button>
                <p className="auth-footer">
                    Don't have an account?
                    <Link href="/register" >
                        Create one
                    </Link>
                </p>
            </form>

        </section >

    )


}