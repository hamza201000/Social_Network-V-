"use client";
import CardPost from "@/components/post/cardPost";
import "../globals.css"
import { PostData } from "@/types/postData";
import CreatePost from "@/components/post/createPost";
import { useEffect, useState } from "react";
import { log } from "util";
import Post from "@/components/post/post";



export default function Home() {

  return (
    <>
      <Post query={undefined} />;
    </>
    )
}
