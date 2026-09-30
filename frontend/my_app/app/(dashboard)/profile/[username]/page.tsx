import Profile from "@/components/profile/profile";
import { useParams } from "next/dist/client/components/navigation";


export default async function UserProfilePage({params}: {params: {username: string}}) {
    const {username} = await params;
    return <Profile username={username} myProfile={false} />;
}