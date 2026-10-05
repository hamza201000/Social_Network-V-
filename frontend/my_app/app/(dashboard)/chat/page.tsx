
"use client";

import { FormEvent, useMemo, useState } from "react";
import "./chat.css";

type Message = {
    id: number;
    text: string;
    time: string;
    fromMe: boolean;
};

type Conversation = {
    id: string;
    name: string;
    kind: "person" | "group";
    initials: string;
    avatarTone: string;
    status: string;
    preview: string;
    time: string;
    unread?: number;
    messages: Message[];
};

const initialConversations: Conversation[] = [];

type IconName = "search" | "more" | "phone" | "video" | "send" | "smile" | "clip" | "back";

function Icon({ name }: { name: IconName }) {
    const glyphs: Record<IconName, string> = {
        search: "⌕",
        more: "⋯",
        phone: "☎",
        video: "▣",
        send: "➤",
        smile: "☺",
        clip: "+",
        back: "‹",
    };
    return <span className={`chat-icon chat-icon-${name}`} aria-hidden="true">{glyphs[name]}</span>;
}

function Avatar({ conversation, small = false }: { conversation: Conversation; small?: boolean }) {
    return (
        <span className={`chat-avatar tone-${conversation.avatarTone}${small ? " chat-avatar-small" : ""}`} aria-hidden="true">
            {conversation.initials}
            {conversation.kind === "person" && conversation.status === "Online now" && <i className="avatar-presence" />}
        </span>
    );
}

export default function ChatPage() {
    const [conversations, setConversations] = useState(initialConversations);
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [filter, setFilter] = useState<"all" | "people" | "groups">("all");
    const [search, setSearch] = useState("");
    const [draft, setDraft] = useState("");
    const [mobileChatOpen, setMobileChatOpen] = useState(false);

    const selected = conversations.find((conversation) => conversation.id === selectedId) ?? null;
    const visibleConversations = useMemo(() => conversations.filter((conversation) => {
        const matchesFilter = filter === "all" || (filter === "groups" ? conversation.kind === "group" : conversation.kind === "person");
        return matchesFilter && conversation.name.toLowerCase().includes(search.toLowerCase());
    }), [conversations, filter, search]);

    function selectConversation(id: string) {
        setSelectedId(id);
        setMobileChatOpen(true);
        setConversations((current) => current.map((conversation) => conversation.id === id ? { ...conversation, unread: undefined } : conversation));
    }

    function sendMessage(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const text = draft.trim();
        if (!text) return;
        if (!selectedId) return;
        setConversations((current) => current.map((conversation) => conversation.id === selectedId ? {
            ...conversation,
            preview: text,
            time: "Now",
            messages: [...conversation.messages, { id: Date.now(), text, time: "Now", fromMe: true }],
        } : conversation));
        setDraft("");
    }

    return (
        <main className="chat-page">
            <section className={`chat-shell${mobileChatOpen ? " mobile-chat-open" : ""}`} aria-label="Messages">
                <aside className="conversation-sidebar">
                    <header className="inbox-header">
                        <div>
                            <span className="eyebrow">YOUR SPACE</span>
                            <h1>Messages<span className="inbox-count">{conversations.length}</span></h1>
                        </div>
                        <button className="icon-button compose-button" type="button" aria-label="Start a new conversation" title="New message">+</button>
                    </header>

                    <label className="conversation-search">
                        <Icon name="search" />
                        <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search conversations" aria-label="Search conversations" />
                        <kbd>⌘ K</kbd>
                    </label>

                    <div className="conversation-filters" role="tablist" aria-label="Conversation type">
                        {(["all", "people", "groups"] as const).map((item) => (
                            <button key={item} type="button" role="tab" aria-selected={filter === item} className={filter === item ? "filter-active" : ""} onClick={() => setFilter(item)}>
                                {item === "all" ? "All" : item === "people" ? "People" : "Groups"}
                            </button>
                        ))}
                    </div>

                    <div className="conversation-list">
                        <div className="list-label">RECENT</div>
                        {visibleConversations.length ? visibleConversations.map((conversation) => (
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
                        )) : <p className="empty-search">{conversations.length ? "No conversations found." : "No conversations yet."}</p>}
                    </div>

                    <footer className="sidebar-footer"><span className="footer-status-dot" /> All caught up <span>·</span> {conversations.length} conversations</footer>
                </aside>

                <section className="chat-conversation" aria-label={selected ? `Conversation with ${selected.name}` : "Conversation panel"}>
                    {selected ? <>
                    <header className="chat-header">
                        <button className="icon-button mobile-back" type="button" aria-label="Back to conversations" onClick={() => setMobileChatOpen(false)}><Icon name="back" /></button>
                        <Avatar conversation={selected} />
                        <div className="chat-header-copy">
                            <h2>{selected.name}</h2>
                            <p><span className={selected.status === "Online now" ? "online-dot" : "status-dot"} />{selected.status}</p>
                        </div>
                        <div className="chat-actions">
                            <button className="icon-button" type="button" aria-label="Start audio call" title="Audio call"><Icon name="phone" /></button>
                            <button className="icon-button" type="button" aria-label="Start video call" title="Video call"><Icon name="video" /></button>
                            <span className="action-divider" />
                            <button className="icon-button" type="button" aria-label="More conversation options" title="More options"><Icon name="more" /></button>
                        </div>
                    </header>

                    <div className="message-area" aria-live="polite">
                        <div className="conversation-date"><span />TODAY<span /></div>
                        <div className="message-stack">
                            {selected.messages.map((message, index) => {
                                const previous = selected.messages[index - 1];
                                const showAvatar = !message.fromMe && (previous?.fromMe || !previous);
                                return (
                                    <div key={message.id} className={`message-row${message.fromMe ? " message-row-mine" : ""}`}>
                                        {!message.fromMe && <span className="message-avatar-slot">{showAvatar ? <Avatar conversation={selected} small /> : null}</span>}
                                        <div className={`message-content${message.fromMe ? " message-content-mine" : ""}`}>
                                            {selected.kind === "group" && !message.fromMe && showAvatar && <span className="message-sender">{message.id === 1 ? "Maya Chen" : "Noah Williams"}</span>}
                                            <p className="message-bubble">{message.text}</p>
                                            <time className="message-time">{message.time}{message.fromMe && <span className="read-check">✓✓</span>}</time>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    <form className="message-composer" onSubmit={sendMessage}>
                        <div className="composer-tools">
                            <button className="icon-button" type="button" aria-label="Attach a file" title="Attach file"><Icon name="clip" /></button>
                            <button className="icon-button" type="button" aria-label="Add emoji" title="Add emoji"><Icon name="smile" /></button>
                        </div>
                        <input value={draft} onChange={(event) => setDraft(event.target.value)} placeholder={`Message ${selected.kind === "group" ? "the group" : selected.name.split(" ")[0]}...`} aria-label="Write a message" />
                        <button className="send-button" type="submit" aria-label="Send message" disabled={!draft.trim()}><Icon name="send" /></button>
                    </form>
                    <p className="composer-note">Enter to send <span>·</span> Messages are a local preview</p>
                    </> : <div className="chat-empty-state"><span className="empty-chat-icon"><Icon name="send" /></span><h2>No conversation selected</h2><p>Choose a conversation to view messages.</p></div>}
                </section>
            </section>
        </main>
    );
}