import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { UserBadge } from "@/lib/types/wire.gen";
import { cn, formatNumber } from "@/lib/utils";
import { useId } from "react";
import { badgeIcons, CreatorTier } from "./badgeIcons";

type BadgeInfo = {
  svg: (prefix: string) => string;
  title: string;
  description: string;
};

// badgeInfo is how a badge of a user looks and what its tooltip says, or
// undefined for badges this version of the site doesn't know.
function badgeInfo(badge: UserBadge): BadgeInfo | undefined {
  switch (badge.kind) {
    case "contributor":
      return {
        svg: badgeIcons.contributor,
        title: "Contributor",
        description: "Helped build Kite",
      };
    case "early_supporter":
      return {
        svg: badgeIcons.early_supporter,
        title: "Early supporter",
        description: "Paid for Kite before Kite Plus launched in October 2026",
      };
    case "subscriber":
      return {
        svg: badgeIcons.subscriber,
        title: badge.plan ? `Kite ${badge.plan} subscriber` : "Subscriber",
        description: "Supports Kite with a paid plan",
      };
    case "early_creator":
      return {
        svg: badgeIcons.early_creator,
        title: "Early creator",
        description:
          "Shared a template in the Marketplace before November 13, 2026",
      };
    case "creator": {
      const tier = badge.milestone as CreatorTier;
      if (![10, 100, 1000].includes(tier)) return undefined;
      return {
        svg: (prefix) => badgeIcons.creator(prefix, tier),
        title: "Template creator",
        description: `${formatNumber(tier)}+ other users added their templates`,
      };
    }
  }
}

// UserBadges shows the badges of a user in a row, like on a Discord profile,
// with what each one means on hover.
export function UserBadges({
  badges,
  className,
  size = "md",
}: {
  badges: UserBadge[];
  className?: string;
  size?: "sm" | "md";
}) {
  const known = badges.flatMap((b) => {
    const info = badgeInfo(b);
    return info ? [{ badge: b, info }] : [];
  });
  if (!known.length) {
    return null;
  }

  return (
    <div
      className={cn(
        "inline-flex items-center rounded-lg bg-muted",
        size === "md" ? "gap-1.5 px-2 py-1.5" : "gap-1 px-1.5 py-1",
        className
      )}
    >
      {known.map(({ badge, info }) => (
        <Tooltip key={badge.kind}>
          <TooltipTrigger asChild>
            <span tabIndex={0} aria-label={info.title} className="flex">
              <BadgeIcon
                svg={info.svg}
                className={size === "md" ? "h-6 w-6" : "h-5 w-5"}
              />
            </span>
          </TooltipTrigger>
          <TooltipContent className="text-center">
            <div className="font-semibold">{info.title}</div>
            <div className="text-xs text-muted-foreground">
              {info.description}
            </div>
          </TooltipContent>
        </Tooltip>
      ))}
    </div>
  );
}

// StaffPickIcon marks a template the Kite team picked.
export function StaffPickIcon({ className }: { className?: string }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span tabIndex={0} aria-label="Staff pick" className="flex">
          <BadgeIcon svg={badgeIcons.staff_pick} className={className} />
        </span>
      </TooltipTrigger>
      <TooltipContent className="text-center">
        <div className="font-semibold">Staff pick</div>
        <div className="text-xs text-muted-foreground">
          Picked by the Kite team
        </div>
      </TooltipContent>
    </Tooltip>
  );
}

function BadgeIcon({
  svg,
  className,
}: {
  svg: (prefix: string) => string;
  className?: string;
}) {
  // useId has colons, which IDs in url() references can't have.
  const prefix = "badge" + useId().replace(/[^a-zA-Z0-9]/g, "");
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden
      className={cn("flex-none", className)}
      dangerouslySetInnerHTML={{ __html: svg(prefix) }}
    />
  );
}
