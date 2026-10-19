import { ReactNode, useRef } from "react";
import HomeNavbar from "./HomeNavbar";
import BaseLayout from "../common/BaseLayout";
import { useScrollReveal } from "@/lib/hooks/scrollReveal";

export default function HomeLayout({
  children,
  scrollAnimations,
  ...props
}: {
  children: ReactNode;
  // Animates the sections in as they scroll into view. Only for the homepage;
  // other pages reuse some of its sections but shouldn't animate.
  scrollAnimations?: boolean;
  title?: string;
  description?: string;
  noindex?: boolean;
  jsonLd?: object[];
}) {
  const rootRef = useRef<HTMLDivElement>(null);
  useScrollReveal(rootRef, !!scrollAnimations);

  return (
    <BaseLayout {...props}>
      {/* overflow-x-clip instead of overflow-hidden, so the sticky navbar works */}
      <div
        ref={rootRef}
        data-reveal-root={scrollAnimations ? "" : undefined}
        className="min-h-[100dvh] flex flex-col overflow-x-clip"
      >
        {scrollAnimations && (
          <noscript>
            <style>{`[data-reveal-root] * { opacity: 1 !important; }`}</style>
          </noscript>
        )}
        <HomeNavbar />
        <div className="flex-auto overflow-hidden">{children}</div>
      </div>
    </BaseLayout>
  );
}
