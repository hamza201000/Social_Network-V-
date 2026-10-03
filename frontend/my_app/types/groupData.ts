export type GroupRequest = {
  id: string;
  username: string;
};

export type Group = {
  id: string;
  name: string;
  description: string;
  category: string;
  creator: string;
  memberCount: number;
  color: string;
  requests: GroupRequest[];
  invitedUsers: string[];
};
