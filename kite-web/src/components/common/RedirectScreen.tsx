import { LoaderCircleIcon } from "lucide-react";
import { useEffect, useState } from "react";
import BaseLayout from "./BaseLayout";

// RedirectScreen sends the browser to url, like Discord's login page, and
// shows a spinner meanwhile. If nothing happened after a few seconds, a link
// shows in case the redirect was blocked.
export default function RedirectScreen({
  title,
  url,
  linkLabel,
  noindex = false,
}: {
  title: string;
  url?: string;
  linkLabel: string;
  noindex?: boolean;
}) {
  const [slow, setSlow] = useState(false);

  useEffect(() => {
    if (!url) return;
    window.location.replace(url);
    const timer = setTimeout(() => setSlow(true), 5000);
    return () => clearTimeout(timer);
  }, [url]);

  return (
    <BaseLayout title={title} noindex={noindex}>
      <div className="flex flex-col flex-1 justify-center items-center gap-4 min-h-[100dvh] w-full px-5 pt-10 pb-20">
        <LoaderCircleIcon
          className="h-8 w-8 animate-spin text-muted-foreground"
          aria-label="Loading"
        />
        {slow && url && (
          <a
            href={url}
            className="text-sm text-muted-foreground hover:text-foreground underline"
          >
            {linkLabel}
          </a>
        )}
      </div>
    </BaseLayout>
  );
}
