import RedirectScreen from "@/components/common/RedirectScreen";
import { installUrl } from "@/lib/api/client";
import { useEffect, useState } from "react";

// The short link for adding the Kite bot, like in links from Xenon. The query
// is passed on, so kite.onl/add?guild_id=…&templates=verification picks the
// server and the templates.
export default function AddPage() {
  // The query is only known in the browser.
  const [url, setUrl] = useState<string>();
  useEffect(() => {
    setUrl(installUrl(window.location.search.slice(1)));
  }, []);

  return (
    <RedirectScreen
      noindex
      title="Add the Kite bot"
      url={url}
      linkLabel="Add the Kite bot with Discord"
    />
  );
}
