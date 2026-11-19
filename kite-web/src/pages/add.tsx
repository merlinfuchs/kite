import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import { installUrl } from "@/lib/api/client";
import { useEffect, useState } from "react";

// The short link for adding the Kite bot, like in links from Xenon. The query
// is passed on, so kite.onl/add?guild_id=…&templates=verification picks the
// server and the templates.
export default function AddPage() {
  const [url, setUrl] = useState<string>();

  useEffect(() => {
    const target = installUrl(window.location.search.slice(1));
    setUrl(target);
    window.location.replace(target);
  }, []);

  return (
    <BaseLayout title="Add the Kite bot" noindex>
      <div className="flex flex-1 justify-center items-center min-h-[100dvh] w-full px-5 pt-10 pb-20">
        {url && (
          <Button asChild>
            <a href={url}>Add the Kite bot with Discord</a>
          </Button>
        )}
      </div>
    </BaseLayout>
  );
}
