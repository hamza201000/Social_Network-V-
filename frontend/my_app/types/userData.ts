export type UserData = {
  user_id: number;
  first_name: string;
  last_name: string;
  username: string;
  email: string;
  avatar_url: string;
  date_of_birth: string;
  about_me: string;
  followed_by_me:boolean;
  followers_count?:number
    following_count?:number
  created_at: string;
};


export type ProfileConnection={
  user_id: number;
  username: string;
    avatarUrl?: string;
}