import { useUpsellStateStore } from "@/lib/state/upsellPopup";
import { XIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "../ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from "../ui/card";
import { useAppId } from "@/lib/hooks/params";
import env from "@/lib/env/client";

export default function UpsellPopup() {
  const shouldUpsell = useUpsellStateStore((s) => s.shouldUpsell);
  const setUpsellClosed = useUpsellStateStore((s) => s.setUpsellClosed);

  const [showUpsell, setShowUpsell] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      setShowUpsell(shouldUpsell());
    }, 5000);

    return () => {
      clearInterval(interval);
    };
  }, [shouldUpsell]);

  const appId = useAppId();

  if (!showUpsell) {
    return null;
  }

  return (
    <Card className="shadow-md max-w-96 fixed bottom-5 right-5 ml-5 z-50">
      <CardHeader className="px-5 py-4">
        <CardTitle className="text-lg">Kite ♥️</CardTitle>
        <CardDescription>
          This is an open source project and almost entirely free to use. If you
          like it,{" "}
          <span className="text-foreground">
            consider supporting the project by getting premium
          </span>{" "}
          or starring the project on Github.
        </CardDescription>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-0 right-1"
          onClick={() => setUpsellClosed(true)}
        >
          <XIcon className="w-5 h-5 cursor-pointer" />
        </Button>
      </CardHeader>
      <CardFooter className="px-5 pb-4 flex gap-3">
        <Button asChild>
          <Link href={`/apps/${appId}/premium`} className="w-full">
            Get Premium
          </Link>
        </Button>
        <Button variant="outline" asChild>
          <Link href={env.NEXT_PUBLIC_GITHUB_LINK} target="_blank">
            Star on Github
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
