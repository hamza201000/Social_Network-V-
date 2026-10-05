import type { Group } from "@/types/groupData";

export default function SmallCard({ groupData }: { groupData: Group }) {
    return (
        <article className="group-search-result">
            <span
                className="group-search-result-mark"
                style={{ backgroundColor: groupData.color }}
                aria-hidden="true"
            >
                {groupData.title.trim().slice(0, 1).toUpperCase() || "G"}
            </span>
            <div className="group-search-result-copy">
                <h3>{groupData.title}</h3>
                <p>{groupData.description}</p>
                <div className="group-search-result-meta">
                    <span>{groupData.memberCount} members</span>
                    <span aria-hidden="true">·</span>
                    <span>by {groupData.admin_name}</span>
                </div>
            </div>
            <span className="group-search-result-arrow" aria-hidden="true">↗</span>
        </article>
    );
}