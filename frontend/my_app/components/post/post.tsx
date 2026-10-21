"use client";
import CardPost from "@/components/post/cardPost";
import "@/app/globals.css"
import { PostData } from "@/types/postData";
import CreatePost from "@/components/post/createPost";
import { useEffect, useState } from "react";
import { API_BASE_URL } from "@/lib/api";




export default function Post({ query, myposts,group_id }: { query: string | undefined, myposts: boolean,group_id:string |undefined}) {

    const [posts, setPosts] = useState<PostData[]>([]);
    let apiKey = query ? (myposts ? `/api/myposts` : `/api/posts/${query}`) : "/api/posts";
    group_id && (apiKey = `/api/groups/${group_id}/posts`);
    const fetchPosts = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}${apiKey}`, {
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
            {!query && !group_id && <CreatePost onPostCreated={fetchPosts} />}
            
                {posts?.length > 0 ? posts.map((post) => (
                    <CardPost key={post.id} postData={post} />
                )) : (
                    <div className="post">
                        <p>No posts available.</p>
                    </div>
                )}
            
        </div>
    );
}
