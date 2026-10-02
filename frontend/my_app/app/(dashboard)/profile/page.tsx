
import Profile from "@/components/profile/profile";

export default async function ProfilePage() {
    return <Profile username="current_user" myProfile={true} />;
}