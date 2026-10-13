import { useRouter } from "next/router";
import SupportDashboard from "@/components/support/SupportDashboard";
import { LoaderCircleIcon } from "lucide-react";

export default function SupportAppPage() {
  const router = useRouter();
  const appId = router.query.appId;

  if (!router.isReady || typeof appId !== "string") {
    return (
      <main className="grid min-h-screen place-items-center">
        <LoaderCircleIcon className="size-6 animate-spin text-muted-foreground" />
      </main>
    );
  }

  return <SupportDashboard appId={appId} />;
}
