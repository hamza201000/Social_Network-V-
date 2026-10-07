import { ProfileConnection } from "@/types/userData";


export default function ConnectedProfiles({
    connection,
    isSelected = false,
    onSelect,
}: {
    connection: ProfileConnection;
    isSelected?: boolean;
    onSelect?: (userId: number) => void;
}) {
    return (
        <li className="connection-item">
            {onSelect ? (
                <button
                    type="button"
                    className="connection-select-button"
                    aria-pressed={isSelected}
                    onClick={() => onSelect(connection.user_id)}
                >
                    <span className="connection-avatar" aria-hidden="true">
                        {connection.avatarUrl ? (
                            <img src={connection.avatarUrl} alt="" />
                        ) : (
                            <span>{connection.username.slice(0, 1).toUpperCase()}</span>
                        )}
                    </span>
                    <span className="connection-username">{connection.username}</span>
                    <span className="connection-selection-mark" aria-hidden="true">
                        {isSelected ? "Selected" : "Select"}
                    </span>
                </button>
            ) : (
                <>
                    <div className="connection-avatar" aria-hidden="true">
                        {connection.avatarUrl ? (
                            <img src={connection.avatarUrl} alt="" />
                        ) : (
                            <span>{connection.username.slice(0, 1).toUpperCase()}</span>
                        )}
                    </div>
                    <span className="connection-username">{connection.username}</span>
                </>
            )}
        </li>
    )
}