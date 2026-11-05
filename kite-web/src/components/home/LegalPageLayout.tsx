import { cn } from "@/lib/utils";
import { formatLegalDate } from "@/lib/legal";
import env from "@/lib/env/client";
import { CalendarIcon, ChevronDownIcon, MailIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { ReactNode, useEffect, useRef, useState } from "react";
import HomeFooter from "./HomeFooter";
import HomeNavbar from "./HomeNavbar";
import BaseLayout from "../common/BaseLayout";

const legalPages = [
  { href: "/terms", label: "Terms of Service" },
  { href: "/privacy", label: "Privacy Policy" },
  { href: "/refund", label: "Refund Policy" },
  { href: "/imprint", label: "Legal Notice" },
];

interface Section {
  id: string;
  title: string;
}

// LegalPageLayout is the layout of the terms, privacy policy, refund policy
// and legal notice. The table of contents is read from the h2s with an id in
// the page's content, so it can't get out of sync with the text.
export default function LegalPageLayout({
  title,
  heading,
  updatedAt,
  noindex,
  children,
}: {
  title: string;
  /** Shown instead of the title on the page, when it should differ. */
  heading?: string;
  updatedAt?: Date;
  noindex?: boolean;
  children: ReactNode;
}) {
  const router = useRouter();
  const contentRef = useRef<HTMLDivElement>(null);
  const tabsRef = useRef<HTMLElement>(null);
  const [sections, setSections] = useState<Section[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    const headings = Array.from(
      contentRef.current?.querySelectorAll<HTMLHeadingElement>("h2[id]") ?? []
    );
    setSections(
      headings.map((h) => ({ id: h.id, title: h.textContent ?? h.id }))
    );

    // The active section is the last heading that scrolled past the top.
    function onScroll() {
      let current: string | null = headings[0]?.id ?? null;
      for (const h of headings) {
        if (h.getBoundingClientRect().top <= 120) current = h.id;
      }
      setActiveId(current);
    }
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [router.pathname]);

  // On small screens the tabs scroll sideways, so bring the active one in.
  useEffect(() => {
    const nav = tabsRef.current;
    const active = nav?.querySelector<HTMLElement>('[aria-current="page"]');
    if (nav && active) {
      nav.scrollLeft =
        active.offsetLeft -
        nav.offsetLeft -
        (nav.clientWidth - active.clientWidth) / 2;
    }
  }, [router.pathname]);

  const toc = (
    <ul className="space-y-1 text-sm">
      {sections.map((s) => (
        <li key={s.id}>
          <a
            href={`#${s.id}`}
            className={cn(
              "block border-l-2 py-1 pl-3 pr-2 transition-colors",
              activeId === s.id
                ? "border-primary font-medium text-foreground"
                : "border-transparent text-muted-foreground hover:border-border hover:text-foreground"
            )}
          >
            {s.title}
          </a>
        </li>
      ))}
    </ul>
  );

  return (
    // Not HomeLayout, as its overflow-hidden wrapper would break the sticky
    // table of contents.
    <BaseLayout title={title} noindex={noindex}>
      <div className="flex min-h-[100dvh] flex-col overflow-x-clip">
        <HomeNavbar />
        <section className="border-b bg-muted/30">
          <div className="container max-w-6xl px-5 py-12 md:py-16">
            <p className="text-sm font-medium text-primary">Legal</p>
            <h1 className="mt-2 text-3xl font-bold tracking-tight md:text-5xl">
              {heading ?? title}
            </h1>
            {updatedAt && (
              <p className="mt-4 inline-flex items-center gap-2 text-sm text-muted-foreground">
                <CalendarIcon className="h-4 w-4" />
                Last updated {formatLegalDate(updatedAt)}
              </p>
            )}

            <nav
              ref={tabsRef}
              aria-label="Legal pages"
              className="-mx-1 mt-8 flex gap-1 overflow-x-auto pb-1"
            >
              {legalPages.map((page) => {
                const active = router.pathname === page.href;
                return (
                  <Link
                    key={page.href}
                    href={page.href}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "whitespace-nowrap rounded-full px-3.5 py-1.5 text-sm font-medium transition-colors",
                      active
                        ? "bg-foreground text-background"
                        : "text-muted-foreground hover:bg-accent hover:text-foreground"
                    )}
                  >
                    {page.label}
                  </Link>
                );
              })}
            </nav>
          </div>
        </section>

        <div className="container max-w-6xl px-5 py-10 md:py-14">
          <div className="lg:grid lg:grid-cols-[220px_minmax(0,1fr)] lg:gap-14">
            <aside className="hidden lg:block">
              {sections.length > 0 && (
                <div className="sticky top-24">
                  <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    On this page
                  </p>
                  {toc}
                </div>
              )}
            </aside>

            <div className="min-w-0">
              {sections.length > 0 && (
                <details className="group mb-8 rounded-xl border bg-card lg:hidden">
                  <summary className="flex cursor-pointer list-none items-center justify-between px-4 py-3 text-sm font-medium [&::-webkit-details-marker]:hidden">
                    On this page
                    <ChevronDownIcon className="h-4 w-4 text-muted-foreground transition-transform group-open:rotate-180" />
                  </summary>
                  <div className="border-t px-2 py-3">{toc}</div>
                </details>
              )}

              <div
                ref={contentRef}
                className="prose max-w-3xl text-foreground dark:prose-invert prose-headings:scroll-mt-24 prose-headings:tracking-tight prose-h2:mt-12 prose-h2:border-t prose-h2:pt-8 first:prose-h2:mt-0 prose-a:text-primary prose-a:no-underline hover:prose-a:underline"
              >
                {children}
              </div>

              <div className="mt-14 flex max-w-3xl flex-col gap-4 rounded-2xl border bg-card p-6 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-start gap-4">
                  <div className="flex h-10 w-10 flex-none items-center justify-center rounded-xl bg-primary/10 text-primary">
                    <MailIcon className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-semibold">Questions about this page?</p>
                    <p className="text-sm text-muted-foreground">
                      Email us and we&apos;ll get back to you.
                    </p>
                  </div>
                </div>
                <a
                  href={`mailto:${env.NEXT_PUBLIC_CONTACT_EMAIL}`}
                  className="inline-flex h-9 flex-none items-center justify-center rounded-md border px-4 text-sm font-medium transition-colors hover:bg-accent"
                >
                  {env.NEXT_PUBLIC_CONTACT_EMAIL}
                </a>
              </div>
            </div>
          </div>
        </div>

        <HomeFooter />
      </div>
    </BaseLayout>
  );
}
