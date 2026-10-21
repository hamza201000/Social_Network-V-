export interface ChatMessage {
  id: number;
  type: string;
  username_sender: string;
  sender_id: number;
  receiver_id: number;
  message: string;
  seen: boolean;
  created_at: string;
}

export interface ChatUser {
  user_id: number;
  username: string;
  is_online: boolean;
  avatar_url?: string;
  last_msg: ChatMessage | null;
}

export interface NewMessage {
  receiver_id: number;
  message: string;
}
