import type { LinkProps } from "next/link";
import Router from "next/router";
import type { MouseEvent } from "react";

// The page the user came from without a page load, like
// "/marketplace/categories/[category]", or null after one. _app imports this
// module, so it knows the page before the first route change.
let previousPathname: string | null = null;
let leavingPathname: string | null = null;

if (typeof window !== "undefined") {
  Router.events.on("routeChangeStart", () => {
    leavingPathname = Router.pathname;
  });
  // Only once it's done, since leaving can be cancelled, like in the editor.
  Router.events.on("routeChangeComplete", () => {
    previousPathname = leavingPathname;
  });
}

// The pages a back link can go back to when they're more than the page it
// links to, like the categories of the marketplace.
const backTargets: Record<string, string[]> = {
  "/marketplace": ["/marketplace", "/marketplace/categories/[category]"],
};

// onBackClick is the onClick of a link back to href, like the list of
// templates. Coming from that page, it goes back in history instead, so the
// page keeps its filters and scroll position.
export function onBackClick(href: LinkProps["href"]) {
  const pathname = typeof href === "string" ? href : href.pathname;
  const targets = pathname ? backTargets[pathname] ?? [pathname] : [];
  return (e: MouseEvent) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) {
      return;
    }
    if (!previousPathname || !targets.includes(previousPathname)) return;
    e.preventDefault();
    Router.back();
  };
}
