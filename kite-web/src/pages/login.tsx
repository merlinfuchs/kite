import RedirectScreen from "@/components/common/RedirectScreen";
import { loginUrl } from "@/lib/api/client";

export default function LoginPage() {
  return (
    <RedirectScreen
      title="Login"
      url={loginUrl}
      linkLabel="Log in with Discord"
    />
  );
}
