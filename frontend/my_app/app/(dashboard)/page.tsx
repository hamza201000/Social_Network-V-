"use client";
import CardPost from "@/components/post/cardPost";
import "../globals.css"
import { PostData } from "@/types/postData";
import CreatePost from "@/components/post/createPost";
import { useEffect, useState } from "react";
import { log } from "util";



 export default function Home() {

  const [posts, setPosts] = useState<PostData[]>([]);

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
  useEffect(() => {
    fetchPosts();
  }, []);
  return (

    <div className="feed">
      <CreatePost onPostCreated={fetchPosts} />
      {posts.length > 0 && posts.map((post) => (
        <CardPost key={post.id} postData={post} />
      ))}
    </div>
  );
}
