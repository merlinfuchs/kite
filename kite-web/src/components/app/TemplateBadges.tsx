import { CreatorBadge, earlyCreatorCutoff } from "@/lib/flow/creatorBadges";
import { cn, formatNumber } from "@/lib/utils";
import { AwardIcon, SparklesIcon, UsersIcon } from "lucide-react";

const cutoffDate = new Date(earlyCreatorCutoff).toLocaleDateString("en-US", {
  month: "long",
  day: "numeric",
  year: "numeric",
  timeZone: "UTC",
});

export function CreatorBadges({ badges }: { badges: CreatorBadge[] }) {
  if (!badges.length) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2">
      {badges.map((badge) =>
        badge.kind === "early" ? (
          <Badge
            key="early"
            icon={SparklesIcon}
            title={`Shared a template in the Marketplace before ${cutoffDate}`}
          >
            Early creator
          </Badge>
        ) : (
          <Badge
            key="milestone"
            icon={UsersIcon}
            title={`Their templates were added by ${formatNumber(
              badge.installers
            )} or more other users`}
          >
            Added by {formatNumber(badge.installers)}+ users
          </Badge>
        )
      )}
    </div>
  );
}

export function StaffPickBadge({ className }: { className?: string }) {
  return (
    <Badge
      icon={AwardIcon}
      title="Picked by the Kite team"
      className={className}
    >
      Staff pick
    </Badge>
  );
}

function Badge({
  icon: Icon,
  title,
  className,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <span
      title={title}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border border-primary/30 bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary",
        className
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {children}
    </span>
  );
}
