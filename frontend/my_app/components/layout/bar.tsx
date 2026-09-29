import Link from "next/link";


export default function Bar() {
    return (

        <aside className="sidebar-top">
            <nav className="nav-menu">
                <Link href="/" className="nav-item active">🏠 Feed</Link>
                <Link href="/profile" className="nav-item">👤 Profile</Link>
                <Link href="/group" className="nav-item">👨‍👩‍👧‍👦 Groups</Link>
                <Link href="/chat" className="nav-item">💬 Chat</Link>
                <Link href="/Settings" className="nav-item">⚙️ Settings</Link>
            </nav>
        </aside>
    )


}