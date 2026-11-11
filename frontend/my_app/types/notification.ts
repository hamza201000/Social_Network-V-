export type NotificationType =
  | "new_follower"
  | "group_invite"
  | "join_request"
  | "group_event"
  | "invite_accepted"
  | "invite_declined"
  | "join_accepted"
  | "join_refused";

export type AppNotification = {
  id: number;
  user_id: number;
  actor_id: number;
  actor_username: string;
  actor_avatar: string;
  type: NotificationType;
  group_id: number | null;
  group_title: string;
  event_id: number | null;
  message: string;
  is_read: boolean;
  created_at: string;
};

export type NotificationList = {
  notifications: AppNotification[];
  unread_count: number;
};

// shape of a notification pushed over the websocket
export type NotificationPush = {
  type: "notification";
  notification: AppNotification;
};
