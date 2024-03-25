import { getApiUrl } from "@/lib/api/client";
import AppIllustrationPlaceholder from "./app/AppIllustrationPlaceholder";
import { Button } from "./ui/button";
import { useEffect } from "react";

export default function LoginPrompt() {
  useEffect(() => {
    setTimeout(() => {
      location.href = getApiUrl("/v1/auth/redirect");
    }, 250);
  }, []);

  return (
    <div className="my-32 lg:my-48 px-5">
      <div className="text-center">
        <Button size="lg" asChild>
          <a href={getApiUrl("/v1/auth/redirect")}>Login with Discord</a>
        </Button>
      </div>
    </div>
  );
}
