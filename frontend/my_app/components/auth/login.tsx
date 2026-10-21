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
        setLoading(true)
        event.preventDefault();
        const Form = new FormData()
        Form.append("username", username)
        Form.append("password", password)
        const body = { username, password,email:username }
        try {
            const response = await fetch(`${API_BASE_URL}/api/login`, {
                method: "POST",
                headers: { "content-type": "application/json" },
                body: JSON.stringify(body),
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
            setError(error)
        } finally {
            router.replace("/")
            router.refresh()
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
                    Sign in
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