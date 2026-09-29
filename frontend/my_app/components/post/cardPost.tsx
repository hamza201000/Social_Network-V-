
import "./post.css"
import { PostData } from "@/types/postData"
export default function CardPost({ postData }: { postData: PostData }) {

    return (
        <article className="post">
            <h3>{postData.content}</h3>
            <header>
                <h2>{postData.username}</h2>
                <time dateTime={postData.created_at}>
                    {postData.created_at}
                </time>
            </header>
            <div>
                <p>{postData.content}</p>

                {postData.image && (
                    <img
                        src={postData.image}
                        alt={`Image posted by ${postData.username}`}
                    />
                )}
            </div>
        </article>



    )




}