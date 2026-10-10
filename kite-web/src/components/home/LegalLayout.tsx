import HomeFooter from "@/components/home/HomeFooter";
import HomeNavbar from "@/components/home/HomeNavbar";
import BaseLayout from "@/components/common/BaseLayout";
import { cn } from "@/lib/utils";
import { ReactNode, useEffect, useRef, useState } from "react";

interface Heading {
  id: string;
  text: string;
}

/**
 * Layout of the legal pages. The text is kept at a readable width, and the
 * table of contents is built from the `h2` headings of the page.
 */
export default function LegalLayout({
  title,
  heading,
  updated,
  children,
}: {
  /** Title of the browser tab. */
  title: string;
  /** Heading shown at the top of the page. */
  heading: string;
  /** Date the text was last changed, like "July 27, 2024". */
  updated: string;
  children: ReactNode;
}) {
  const articleRef = useRef<HTMLElement>(null);
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [active, setActive] = useState<string | null>(null);

  useEffect(() => {
    const article = articleRef.current;
    if (!article) return;

    const used = new Set<string>();
    const elements = Array.from(article.querySelectorAll("h2"));

    setHeadings(
      elements.map((el) => {
        const text = el.textContent ?? "";
        let id = slugify(text);
        while (used.has(id)) id += "-";
        used.add(id);

        el.id = id;
        return { id, text };
      })
    );

    // The last heading that passed the top of the viewport is the section
    // being read.
    let frame = 0;
    function update() {
      frame = 0;
      let current: string | null = null;
      for (const el of elements) {
        if (el.getBoundingClientRect().top > 120) break;
        current = el.id;
      }
      setActive(current);
    }
    function onScroll() {
      if (!frame) frame = requestAnimationFrame(update);
    }

    update();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => {
      window.removeEventListener("scroll", onScroll);
      if (frame) cancelAnimationFrame(frame);
    };
  }, []);

  return (
    // HomeLayout hides its overflow, which would keep the table of contents
    // from sticking, so the navbar is put together here instead.
    <BaseLayout title={title}>
      <HomeNavbar />
      {/* The table of contents sits at the left edge of the page, in line
          with the navbar. From xl on an empty column of the same width on
          the right keeps the text in the middle of the page. */}
      <div className="px-5 my-12 md:my-16 lg:grid lg:grid-cols-[15rem,minmax(0,1fr)] lg:gap-10 xl:grid-cols-[15rem,minmax(0,1fr),15rem]">
        <aside className="hidden lg:block">
          {headings.length > 0 ? (
            <nav
              aria-label="On this page"
              className="sticky top-8 max-h-[calc(100dvh-4rem)] overflow-y-auto no-scrollbar"
            >
              <div className="mb-3 text-sm font-semibold">On this page</div>
              <ul className="space-y-1 border-l text-sm">
                {headings.map((h) => (
                  <li key={h.id}>
                    <a
                      href={`#${h.id}`}
                      className={cn(
                        "-ml-px block border-l py-1 pl-4 transition-colors hover:text-foreground",
                        active === h.id
                          ? "border-primary font-medium text-foreground"
                          : "border-transparent text-muted-foreground"
                      )}
                    >
                      {h.text}
                    </a>
                  </li>
                ))}
              </ul>
            </nav>
          ) : null}
        </aside>

        <article
          ref={articleRef}
          className="prose dark:prose-invert text-foreground mx-auto w-full max-w-3xl prose-headings:scroll-mt-8 prose-h2:mt-12 prose-h2:border-t prose-h2:pt-8"
        >
          <h1 className="mb-3">{heading}</h1>
          <p className="not-prose mb-8 text-sm text-muted-foreground">
            Last updated: {updated}
          </p>
          {children}
        </article>
      </div>
      <HomeFooter />
    </BaseLayout>
  );
}

function slugify(text: string) {
  return (
    text
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "") || "section"
  );
}
