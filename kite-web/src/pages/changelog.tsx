import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { changelog, ChangelogEntry } from "@/lib/changelog";
import { ArrowRightIcon, ChevronDownIcon } from "lucide-react";
import Link from "next/link";

function formatDate(date: string, options: Intl.DateTimeFormatOptions) {
  return new Date(`${date}T00:00:00Z`).toLocaleDateString("en-US", {
    timeZone: "UTC",
    ...options,
  });
}

// Groups items next to each other with the same key, keeping their order.
function groupBy<T>(items: T[], key: (item: T) => string) {
  const groups: { key: string; items: T[] }[] = [];
  for (const item of items) {
    const k = key(item);
    const last = groups[groups.length - 1];
    if (last?.key === k) {
      last.items.push(item);
    } else {
      groups.push({ key: k, items: [item] });
    }
  }
  return groups;
}

// The newest months are open, older ones show how many changes they had
// until they're opened.
const openMonths = 2;

function EntryTitle({ entry }: { entry: ChangelogEntry }) {
  if (!entry.href) return <>{entry.title}</>;
  return (
    <Link href={entry.href} className="hover:text-primary transition-colors">
      {entry.title}
      <ArrowRightIcon className="inline w-4 h-4 ml-1 -mt-0.5" />
    </Link>
  );
}

// One row per day, with the changes that have a description first and the
// small ones in a list below them.
function MonthEntries({ entries }: { entries: ChangelogEntry[] }) {
  return (
    <ol>
      {groupBy(entries, (e) => e.date).map(({ key: date, items }) => {
        const small = items.filter((e) => !e.description);
        return (
          <li
            key={date}
            className="border-t py-5 grid grid-cols-1 sm:grid-cols-[7rem_1fr] gap-x-6 gap-y-2"
          >
            <time
              dateTime={date}
              className="text-sm text-muted-foreground pt-1"
            >
              {formatDate(date, { month: "short", day: "numeric" })}
            </time>
            <div className="space-y-4">
              {items
                .filter((e) => e.description)
                .map((entry) => (
                  <div key={entry.title}>
                    <h3 className="font-semibold text-lg">
                      <EntryTitle entry={entry} />
                    </h3>
                    <p className="text-muted-foreground mt-1">
                      {entry.description}
                    </p>
                  </div>
                ))}
              {small.length > 0 && (
                <ul className="list-disc pl-5 space-y-1 marker:text-muted-foreground">
                  {small.map((entry) => (
                    <li key={entry.title}>
                      <EntryTitle entry={entry} />
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </li>
        );
      })}
    </ol>
  );
}

export default function ChangelogPage() {
  const months = groupBy(changelog, (e) =>
    formatDate(e.date, { month: "long", year: "numeric" })
  );

  return (
    <HomeLayout
      title="Changelog: what's new in Kite"
      description={`New blocks, templates, and plan changes in Kite, the free Discord bot maker. Last update on ${formatDate(
        changelog[0].date,
        { month: "long", day: "numeric", year: "numeric" }
      )}.`}
    >
      <section className="container max-w-3xl py-16 md:py-24">
        <h1 className="text-4xl md:text-5xl font-bold">
          What&apos;s new in Kite
        </h1>
        <p className="text-xl text-muted-foreground mt-6">
          New blocks, templates, and plan changes, newest first. Kite opened to
          everyone in September 2024.
        </p>

        {months.slice(0, openMonths).map(({ key: month, items }) => (
          <div key={month} className="mt-14">
            <h2 className="text-2xl md:text-3xl font-bold mb-4">{month}</h2>
            <MonthEntries entries={items} />
          </div>
        ))}

        <div className="mt-14 border-t">
          {months.slice(openMonths).map(({ key: month, items }) => (
            <details key={month} className="group border-b">
              <summary className="flex items-center gap-3 py-4 cursor-pointer list-none [&::-webkit-details-marker]:hidden">
                <h2 className="text-xl font-bold">{month}</h2>
                <span className="text-sm text-muted-foreground">
                  {items.length} {items.length === 1 ? "change" : "changes"}
                </span>
                <ChevronDownIcon className="w-5 h-5 ml-auto text-muted-foreground transition-transform group-open:rotate-180" />
              </summary>
              <div className="pb-4">
                <MonthEntries entries={items} />
              </div>
            </details>
          ))}
        </div>
      </section>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
