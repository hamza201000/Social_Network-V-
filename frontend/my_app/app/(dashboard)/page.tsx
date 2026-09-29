"use client";
import CardPost from "@/components/post/cardPost";
import "../globals.css"
import { PostData } from "@/types/postData";
import CreatePost from "@/components/post/createPost";
import { useEffect, useState } from "react";



export default function Home() {
  
  const [posts, setPosts] = useState<PostData[]>([]);

  useEffect(() => {
    const fetchPosts = async () => {
      try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/api/posts`, {
          method: "GET",
          credentials: "include"
        });
        if (!response.ok) {
          throw new Error("Failed to fetch posts");
        }
        const data: PostData[] = await response.json();
        setPosts(data);
      } catch (error) {
        console.error("Error fetching posts:", error);
      }
    };
    fetchPosts();
  }, [posts]); 
  const fakePost: PostData[] = [{
  id: 1,
  user_id: 42,
  username: "john_doe",
  content: "Just finished building my first React + TypeScript app! 🚀",
  image: "https://picsum.photos/600/400",
  privacy: "public",
  created_at: "2026-09-29T12:30:00Z",
},{
  id: 2,
  user_id: 42,
  username: "john_doe",
  content: "Just finished building my first React + TypeScript app! 🚀",
  image: "https://picsum.photos/600/400",
  privacy: "public",
  created_at: "2026-09-29T12:30:00Z",
}];
  return (

   <div className="feed">
    <CreatePost />
    {fakePost.map((post) => (
      <CardPost key={post.id} postData={post} />
    ))}
   </div> 
  );
}
