import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { UserBadge } from "@/lib/types/wire.gen";
import { cn, formatNumber } from "@/lib/utils";
import { useId } from "react";
import { badgeIcons, CreatorTier, StarTier } from "./badgeIcons";

type BadgeInfo = {
  svg: (prefix: string) => string;
  title: string;
  description: string;
};

// badgeInfo is how a badge of a user looks and what its tooltip says, or
// undefined for badges this version of the site doesn't know.
function badgeInfo(badge: UserBadge): BadgeInfo | undefined {
  switch (badge.kind) {
    case "staff":
      return {
        svg: badgeIcons.staff,
        title: "Kite staff",
        description: "Part of the team behind Kite",
      };
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
        title: badge.plan ? `${badge.plan} subscriber` : "Subscriber",
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
      const tier = (badge.milestone ?? 0) as CreatorTier;
      if (![0, 10, 100, 1000].includes(tier)) return undefined;
      return {
        svg: (prefix) => badgeIcons.creator(prefix, tier),
        title: tier
          ? `Template creator · ${formatNumber(tier)}+`
          : "Template creator",
        description: tier
          ? `${formatNumber(tier)}+ other users added their templates`
          : "Shared a template in the Marketplace",
      };
    }
    case "star": {
      const tier = (badge.milestone ?? 0) as StarTier;
      if (![5, 20, 50].includes(tier)) return undefined;
      return {
        svg: (prefix) => badgeIcons.star(prefix, tier),
        title: `Star creator · ${formatNumber(tier)}+`,
        description: `Other users gave their templates ${formatNumber(
          tier
        )}+ stars`,
      };
    }
  }
}

// UserBadges shows the badges of a user in a row, like on a Discord profile,
// with what each one means on hover. The labeled version names each badge,
// for profile pages where they're the point.
export function UserBadges({
  badges,
  className,
  size = "md",
  labeled,
}: {
  badges: UserBadge[];
  className?: string;
  size?: "sm" | "md";
  labeled?: boolean;
}) {
  const known = badges.flatMap((b) => {
    const info = badgeInfo(b);
    return info ? [{ badge: b, info }] : [];
  });
  if (!known.length) {
    return null;
  }

  if (labeled) {
    return (
      <div className={cn("flex flex-wrap gap-1.5", className)}>
        {known.map(({ badge, info }) => (
          <Tooltip key={badge.kind}>
            <TooltipTrigger asChild>
              <span
                tabIndex={0}
                className="flex items-center gap-1.5 rounded-full border bg-background py-1 pl-1.5 pr-3 text-sm font-medium"
              >
                <BadgeIcon svg={info.svg} className="h-[22px] w-[22px]" />
                {info.title}
              </span>
            </TooltipTrigger>
            <TooltipContent>{info.description}</TooltipContent>
          </Tooltip>
        ))}
      </div>
    );
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
