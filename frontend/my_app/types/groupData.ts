export type GroupRequest = {
  id: string;
  username: string;
};

export type Group = {
  id: string;
  title: string;
  description: string;
  admin_name: string;
  memberCount: number;
  color: string;
  requests: GroupRequest[];
  invitedUsers: string[];
};
