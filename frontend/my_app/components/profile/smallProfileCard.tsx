import { API_BASE_URL } from "@/lib/api";
import { UserData } from "@/types/userData";
import Link from "next/link";



export default function UserCard({ userData,deleteHandler: deleteHandler }: { userData: UserData,deleteHandler?: () => void }) {
    const fullName = `${userData.first_name} ${userData.last_name}`.trim();

    return (
        <Link href={`/profile/${userData.username}`} className="user-search-result-link" onClick={(e) => {deleteHandler?.()}}>
            <article className="user-search-result">
                <span className="user-search-result-mark">
                    {userData.avatar_url ? (
                        <img
                            src={`${API_BASE_URL}/api/images/${encodeURIComponent(userData.avatar_url)}`}
                        alt={`${fullName}'s avatar`}
                    />
                ) : (
                    fullName.slice(0, 1).toUpperCase() || "U"
                )}
            </span>

            <div className="user-search-result-copy">
                <h3>{fullName || userData.username}</h3>
                <p>@{userData.username}</p>

                {userData.about_me && (
                    <p>{userData.about_me}</p>
                )}

                <div className="user-search-result-meta">
                    <span>{userData.followers_count ?? 0} followers</span>
                    <span aria-hidden="true">·</span>
                    <span>{userData.following_count ?? 0} following</span>
                </div>
            </div>

            <span className="user-search-result-arrow" aria-hidden="true">
                ↗
            </span>
        </article>
    </Link>
);
}