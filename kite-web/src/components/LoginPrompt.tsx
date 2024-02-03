import { getApiUrl } from "@/lib/api/client";
import AppIllustrationPlaceholder from "./app/AppIllustrationPlaceholder";

export default function LoginPrompt() {
  return (
    <div className="my-32 lg:my-48 px-5">
      <AppIllustrationPlaceholder
        className="mb-10"
        svgPath="/illustrations/signin.svg"
        title="Login with your Discord account to access Kite and get started!"
      />
      <div className="text-center">
        <a
          href={getApiUrl("/v1/auth/redirect")}
          className="px-5 py-3 bg-primary rounded text-white text-lg"
        >
          Login with Discord
        </a>
      </div>
    </div>
  );
}
