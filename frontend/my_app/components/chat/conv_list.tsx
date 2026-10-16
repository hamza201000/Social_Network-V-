import { useEffect, useState } from "react";

function Avatar({ conversation, small = false }: { conversation: Conversation; small?: boolean }) {
    return (
        <span className={`chat-avatar tone-${conversation.avatarTone}${small ? " chat-avatar-small" : ""}`} aria-hidden="true">
            {conversation.initials}
            {conversation.kind === "person" && conversation.status === "Online now" && <i className="avatar-presence" />}
        </span>
    );
}

export default function ConversationList({user_id}:{user_id:string}) {
    const [Conversations,setConversations]=useState<Conversation[]>([])
    const [selectedId, setSelectedId] = useState<string | null>(null);

    function selectConversation(id: string) {
        setSelectedId(id);
        setConversations((current) => current.map((conversation) => conversation.id === id ? { ...conversation, unread: undefined } : conversation));
    }
    const fetchConversation = async () =>{
        try{
            const response = await fetch(`http://localhost:8080/api/chat/{user_id}/conversationList`,{
                method:"GET",
                credentials:"include"
            })
            if (!response.ok) {
                throw new Error("Failed to fetch ConversationList");
            }
            const data :Conversation[] = await response.json()
            setConversations(data)
        }catch (err){

        }
    }
    useEffect(()=>{
        fetchConversation()
    },[])

    return (
        <div className="conversation-list">
            <div className="list-label">RECENT</div>
            {Conversations.length ? Conversations.map((conversation) => (
                <button key={conversation.id} type="button" className={`conversation-item${selectedId === conversation.id ? " conversation-active" : ""}`} onClick={() => selectConversation(conversation.id)}>
                    <Avatar conversation={conversation} />
                    <span className="conversation-copy">
                        <span className="conversation-name-row"><strong>{conversation.name}</strong><time>{conversation.time}</time></span>
                        <span className="conversation-preview-row">
                            <span className="conversation-type">{conversation.kind === "group" ? "Group" : conversation.preview}</span>
                            {conversation.unread ? <span className="unread-badge">{conversation.unread}</span> : null}
                        </span>
                        {conversation.kind === "group" && <span className="group-preview">{conversation.preview}</span>}
                    </span>
                </button>
            )) : <p className="empty-search">No conversations</p>}
        </div>
    )
}
