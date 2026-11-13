import type { LinkProps } from "next/link";
import Router from "next/router";
import type { MouseEvent } from "react";

// The pathname of the page each history entry was opened from without a page
// load, like "/marketplace/categories/[category]", by the key Next.js gives
// the entry. Going back from an entry leads to that page. _app imports this
// module, so it sees every route change.
const openedFrom = new Map<string, string>();
let leaving: { pathname: string; key?: string } | null = null;
let popping = false;

const historyKey = (): string | undefined => window.history.state?.key;

if (typeof window !== "undefined") {
  // Going back or forward opens no new entry. Next.js starts the route change
  // after this, since it awaits first.
  window.addEventListener("popstate", () => {
    popping = true;
  });
  Router.events.on("routeChangeStart", () => {
    leaving = popping ? null : { pathname: Router.pathname, key: historyKey() };
    popping = false;
  });
  // Only once it's done, since leaving can be cancelled, like in the editor.
  // Replacing the URL keeps the key of the entry.
  Router.events.on("routeChangeComplete", () => {
    const key = historyKey();
    if (leaving && key && key !== leaving.key) {
      openedFrom.set(key, leaving.pathname);
    }
    leaving = null;
  });
}

// The pages a back link can go back to when they're more than the page it
// links to, like the categories of the marketplace.
const backTargets: Record<string, string[]> = {
  "/marketplace": ["/marketplace", "/marketplace/categories/[category]"],
};

// onBackClick is the onClick of a link back to href, like the list of
// templates. When going back leads to that page, it goes back in history
// instead, so the page keeps its filters and scroll position.
export function onBackClick(href: LinkProps["href"]) {
  const pathname = typeof href === "string" ? href : href.pathname;
  const targets = pathname ? backTargets[pathname] ?? [pathname] : [];
  return (e: MouseEvent) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) {
      return;
    }
    const key = historyKey();
    const previous = key && openedFrom.get(key);
    if (!previous || !targets.includes(previous)) return;
    e.preventDefault();
    Router.back();
  };
}
