import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import { loginUrl } from "@/lib/api/client";
import { useEffect } from "react";

export default function LoginPage() {
  useEffect(() => {
    window.location.href = loginUrl;
  }, []);

  return (
    <BaseLayout title="Login">
      <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
        <Button asChild>
          <a href={loginUrl}>Login with Discord</a>
        </Button>
      </div>
    </BaseLayout>
  );
}
