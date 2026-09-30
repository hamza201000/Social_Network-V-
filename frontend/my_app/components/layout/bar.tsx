"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Bar() {
    const pathname = usePathname();

    const navItems = [
        { href: "/", label: "🏠 Feed" },
        { href: "/profile", label: "👤 Profile" },
        { href: "/group", label: "👨‍👩‍👧‍👦 Groups" },
        { href: "/chat", label: "💬 Chat" },
        { href: "/Settings", label: "⚙️ Settings" },
    ];

    return (
        <aside className="sidebar-top">
            <nav className="nav-menu">
                {navItems.map((item) => (
                    <Link
                        key={item.href}
                        href={item.href}
                        className={`nav-item ${
                            pathname === item.href ? "active" : ""
                        }`}
                    >
                        {item.label}
                    </Link>
                ))}
            </nav>
        </aside>
    );
}
