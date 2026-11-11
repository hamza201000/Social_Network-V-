"use client"

import Link from "next/link";
import { usePathname } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";
import { useRouter } from 'next/navigation';
import { useState } from "react";
import { UserData } from "@/types/userData";
import { SearchUser } from "../profile/searchProfile";
import NotificationBell from "../notification/notificationBell";

export default function Bar() {
    const pathname = usePathname();
    const router = useRouter();
    
    const navItems = [
        { href: "/", label: "🏠 Feed" },
        { href: "/profile", label: "👤 Profile" },
        { href: "/chat", label: "💬 Chat" },
        { href: "/group", label: "👨‍👩‍👧‍👦 Groups" },
    ];
    const handleClick = async () => {
        const response = await fetch(`${API_BASE_URL}/api/logout`, {
            method: "GET",
            credentials: "include",
        });
        if (response.ok) {
            router.replace("/login");
        }
    };
    
    return (
        <aside className="sidebar-top">
            <nav className="nav-menu">
                {navItems.map((item) => (
                    <Link
                        key={item.href}
                        href={item.href}
                        className={`nav-item ${pathname === item.href ? "active" : ""
                            }`}
                    >
                        {item.label}
                    </Link>
                ))}
                <button className="nav-item" onClick={handleClick} >Log out</button>
            </nav>
            <SearchUser />
            <NotificationBell />
        </aside>
    );
}
