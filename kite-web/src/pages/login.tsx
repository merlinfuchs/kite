import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";

export default function LoginPage() {
  const url = env.NEXT_PUBLIC_API_PUBLIC_BASE_URL + "/v1/auth/login";

  return (
    <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
      <Button asChild>
        <a href={url}>Login with Discord</a>
      </Button>
    </div>
  );
}
