
import "./post.css"
import { PostData } from "@/types/postData"
import { imageUrl } from "@/lib/api";
export default function CardPost({ postData }: { postData: PostData }) {

    return (
        <article className="post">
            <header>
                <h2>{postData.username}</h2>
                <time dateTime={postData.created_at}>
                    {postData.created_at}
                </time>
            </header>
            <div>
                <p>{postData.content}</p>

                {imageUrl(postData.image) && (
                    <img
                        src={imageUrl(postData.image)}
                        alt={`Image posted by ${postData.username}`}
                        onError={(event) => {
                            event.currentTarget.hidden = true;
                        }}
                    />
                )}
            </div>
        </article>
    )
}