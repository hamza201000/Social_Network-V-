import CardPost from "@/components/post/cardPost";
import "../globals.css"
import { PostData } from "@/types/postData";


export default function Home() {
  
  
  const fakePost: PostData = {
  id: 1,
  user_id: 42,
  username: "john_doe",
  content: "Just finished building my first React + TypeScript app! 🚀",
  image: "https://picsum.photos/600/400",
  privacy: "public",
  created_at: "2026-09-29T12:30:00Z",
};
  return (
   <div className="feed">
      <CardPost postData={fakePost} />
   </div> 
  );
}
