"use client";
import CardPost from "@/components/post/cardPost";
import "@/app/globals.css"
import { PostData } from "@/types/postData";
import CreatePost from "@/components/post/createPost";
import { useEffect, useState } from "react";
import { log } from "util";



export default function Post({ query }: { query: string | undefined }) {

    const [posts, setPosts] = useState<PostData[]>([]);
    const apiKey= query ? `/api/posts/${query}` : "/api/posts";
    const fetchPosts = async () => {
        try {
            const response = await fetch(`http://localhost:8080${apiKey}`, {
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
    console.log(posts);

    return (

        <div className="feed">
            {!query && <CreatePost onPostCreated={fetchPosts} />}
            {posts.length > 0 && posts.map((post) => (
                <CardPost key={post.id} postData={post} />
            ))}
        </div>
    );
}
