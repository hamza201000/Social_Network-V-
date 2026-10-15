

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