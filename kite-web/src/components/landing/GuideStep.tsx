import { ReactNode } from "react";

// A numbered step of a guide, with its text next to a visual. The steps are
// connected by a dashed line, so the last one needs `last`.
export default function GuideStep({
  number,
  title,
  visual,
  last,
  children,
}: {
  number: number;
  title: string;
  visual: ReactNode;
  last?: boolean;
  children: ReactNode;
}) {
  return (
    <section
      id={`step-${number}`}
      className="relative pl-14 pb-16 scroll-mt-20"
    >
      <div className="absolute left-0 top-0 w-10 h-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
        {number}
      </div>
      {!last && (
        <div className="absolute left-5 top-12 bottom-2 border-l-2 border-dashed" />
      )}
      <h2 className="text-2xl md:text-3xl font-bold leading-10">{title}</h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 mt-4">
        <div className="prose dark:prose-invert text-foreground prose-code:before:content-none prose-code:after:content-none">
          {children}
        </div>
        <div>{visual}</div>
      </div>
    </section>
  );
}

// The table of contents at the top of a guide, linking to its steps.
export function GuideContents({ steps }: { steps: string[] }) {
  return (
    <nav className="mt-12 rounded-lg border bg-muted/50 p-6">
      <div className="font-semibold">In this guide</div>
      <ol className="mt-3 grid sm:grid-cols-2 gap-2">
        {steps.map((title, i) => (
          <li key={title}>
            <a
              href={`#step-${i + 1}`}
              className="text-muted-foreground hover:text-foreground"
            >
              {i + 1}. {title}
            </a>
          </li>
        ))}
      </ol>
    </nav>
  );
}
