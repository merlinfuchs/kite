import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { changelog, ChangelogEntry } from "@/lib/changelog";
import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";

function formatDate(date: string, options: Intl.DateTimeFormatOptions) {
  return new Date(`${date}T00:00:00Z`).toLocaleDateString("en-US", {
    timeZone: "UTC",
    ...options,
  });
}

// The entries grouped by month, newest first like the list.
function byMonth(entries: ChangelogEntry[]) {
  const months: { month: string; entries: ChangelogEntry[] }[] = [];
  for (const entry of entries) {
    const month = formatDate(entry.date, { month: "long", year: "numeric" });
    const last = months[months.length - 1];
    if (last?.month === month) {
      last.entries.push(entry);
    } else {
      months.push({ month, entries: [entry] });
    }
  }
  return months;
}

export default function ChangelogPage() {
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
          New blocks, templates, and plan changes, newest first. Kite started in
          2024, and this list starts in October 2026.
        </p>

        {byMonth(changelog).map(({ month, entries }) => (
          <div key={month} className="mt-14">
            <h2 className="text-2xl md:text-3xl font-bold">{month}</h2>
            <ol className="mt-4">
              {entries.map((entry) => (
                <li
                  key={`${entry.date} ${entry.title}`}
                  className="border-t py-5 grid grid-cols-1 sm:grid-cols-[7rem_1fr] gap-x-6 gap-y-1"
                >
                  <time
                    dateTime={entry.date}
                    className="text-sm text-muted-foreground pt-0.5"
                  >
                    {formatDate(entry.date, { month: "short", day: "numeric" })}
                  </time>
                  <div>
                    <h3 className="font-semibold text-lg">
                      {entry.href ? (
                        <Link
                          href={entry.href}
                          className="inline-flex items-center gap-1 hover:text-primary transition-colors"
                        >
                          {entry.title}
                          <ArrowRightIcon className="w-4 h-4 flex-none" />
                        </Link>
                      ) : (
                        entry.title
                      )}
                    </h3>
                    <p className="text-muted-foreground mt-1">
                      {entry.description}
                    </p>
                  </div>
                </li>
              ))}
            </ol>
          </div>
        ))}
      </section>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
