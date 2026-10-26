import { track } from "@/lib/analytics";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { Button } from "../ui/button";

export default function GetCreditsButton({ from }: { from: string }) {
  const appId = useAppId();

  return (
    <Button size="sm" asChild>
      <Link
        href={`/apps/${appId}/premium`}
        onClick={() => track("upsell_clicked", { from })}
      >
        Get more credits
      </Link>
    </Button>
  );
}
