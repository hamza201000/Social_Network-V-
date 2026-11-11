// The viewer's relation to a group, computed by the backend.
export type GroupStatus = "admin" | "member" | "invited" | "pending" | "none";

export type Group = {
  id: string | number;
  title: string;
  description: string;
  admin_id: number;
  admin_name: string;
  member_count: number;
  my_status: GroupStatus;
};

export type GroupMember = {
  user_id: number;
  first_name: string;
  last_name: string;
  username: string;
  avatar_url: string;
  role?: "admin" | "member";
  created_at: string;
};
