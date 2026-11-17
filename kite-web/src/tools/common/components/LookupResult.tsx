import { Badge } from "@/components/ui/badge";
import CopyButton from "@/tools/common/components/CopyButton";
import { formatRelative } from "@/tools/common/utils/time";
import { ExternalLinkIcon } from "lucide-react";
import { ReactNode } from "react";

export interface LookupRow {
  label: string;
  value: ReactNode;
  // Shows a copy button for this text.
  copy?: string;
}

// The header of a looked up user, server or app, with its banner and icon.
export function LookupHeader({
  bannerUrl,
  bannerColor,
  iconUrl,
  iconDecorationUrl,
  iconShape = "circle",
  title,
  subtitle,
  badges,
}: {
  bannerUrl?: string | null;
  bannerColor?: string | null;
  iconUrl: string | null;
  iconDecorationUrl?: string | null;
  iconShape?: "circle" | "rounded";
  title: ReactNode;
  subtitle?: ReactNode;
  badges?: string[];
}) {
  const shape = iconShape === "circle" ? "rounded-full" : "rounded-2xl";

  return (
    <div className="rounded-lg border overflow-hidden">
      <div
        className="h-28 sm:h-36 bg-muted bg-cover bg-center"
        style={{
          backgroundImage: bannerUrl ? `url(${bannerUrl})` : undefined,
          backgroundColor: bannerColor ?? undefined,
        }}
      />
      <div className="px-5 pb-5">
        <div className="relative -mt-12 w-24 h-24">
          {iconUrl ? (
            <img
              src={iconUrl}
              alt=""
              className={`w-24 h-24 ${shape} border-4 border-card bg-card object-cover`}
            />
          ) : (
            <div
              className={`w-24 h-24 ${shape} border-4 border-card bg-muted`}
            />
          )}
          {iconDecorationUrl && (
            <img
              src={iconDecorationUrl}
              alt=""
              className="absolute inset-0 w-24 h-24 scale-[1.2] pointer-events-none"
            />
          )}
        </div>
        <div className="mt-3 text-2xl font-semibold break-words">{title}</div>
        {subtitle && (
          <div className="text-muted-foreground mt-1 break-words">
            {subtitle}
          </div>
        )}
        {badges && badges.length > 0 && (
          <div className="flex flex-wrap gap-2 mt-3">
            {badges.map((badge) => (
              <Badge key={badge} variant="secondary">
                {badge}
              </Badge>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export function LookupRows({ rows }: { rows: LookupRow[] }) {
  return (
    <div className="rounded-lg border divide-y">
      {rows.map((row) => (
        <div
          key={row.label}
          className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3 p-4"
        >
          <div className="sm:w-48 flex-none font-medium">{row.label}</div>
          <div className="flex-auto text-sm break-all">{row.value}</div>
          {row.copy && <CopyButton value={row.copy} />}
        </div>
      ))}
    </div>
  );
}

export function LookupLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-1 underline"
    >
      {children}
      <ExternalLinkIcon className="w-3 h-3" />
    </a>
  );
}

// Formats when an ID was made, like "March 3, 2021 (4 years ago)".
export function formatCreated(timestamp: number) {
  const date = new Intl.DateTimeFormat(undefined, {
    dateStyle: "long",
    timeStyle: "short",
  }).format(timestamp);
  return `${date} (${formatRelative(Math.floor(timestamp / 1000))})`;
}
