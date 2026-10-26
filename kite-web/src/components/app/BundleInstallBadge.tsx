import { useBundlesQuery } from "@/lib/api/queries";
import { useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { cn } from "@/lib/utils";
import { PackageIcon } from "lucide-react";
import Link from "next/link";
import { badgeVariants } from "../ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

type Props = { bundleInstallId: string | null | undefined };

// useInstalledFrom is the template or bundle an item was added with, or null
// while it loads. Templates the marketplace no longer lists count as bundles,
// their page shows them as gone.
function useInstalledFrom(bundleInstallId: string) {
  const appId = useAppId();
  const install = useBundleInstalls()?.find((i) => i?.id === bundleInstallId);
  const { data: bundles } = useBundlesQuery();
  if (!install || !bundles) return null;

  const official =
    bundles.success &&
    bundles.data.some((b) => b?.id === install.bundle_id && b.official);
  return {
    name: install.name,
    noun: official ? "template" : "bundle",
    href: official
      ? {
          pathname: "/apps/[appId]/marketplace/[templateId]",
          query: { appId, templateId: install.bundle_id },
        }
      : {
          pathname: "/apps/[appId]/bundles/[bundleId]",
          query: { appId, bundleId: install.bundle_id },
        },
  };
}

// BundleInstallBadge links an item in a list to the template or bundle it was
// added with. Items the user made load nothing.
export default function BundleInstallBadge({ bundleInstallId }: Props) {
  return bundleInstallId ? <Badge bundleInstallId={bundleInstallId} /> : null;
}

function Badge({ bundleInstallId }: { bundleInstallId: string }) {
  const from = useInstalledFrom(bundleInstallId);
  if (!from) return null;

  return (
    <Link
      href={from.href}
      className={cn(
        badgeVariants({ variant: "outline" }),
        "gap-1 px-2 font-medium text-muted-foreground hover:text-foreground hover:bg-muted max-w-full"
      )}
      title={`Added with the ${from.name} ${from.noun}`}
    >
      <PackageIcon className="h-3 w-3 flex-none" />
      <span className="truncate">{from.name}</span>
    </Link>
  );
}

// BundleInstallNavItem shows the template or bundle in the nav of an editor.
export function BundleInstallNavItem({
  bundleInstallId,
  className,
}: Props & { className?: string }) {
  return bundleInstallId ? (
    <NavItem bundleInstallId={bundleInstallId} className={className} />
  ) : null;
}

function NavItem({
  bundleInstallId,
  className,
}: {
  bundleInstallId: string;
  className?: string;
}) {
  const from = useInstalledFrom(bundleInstallId);
  if (!from) return null;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          href={from.href}
          aria-label={`Added with the ${from.name} ${from.noun}`}
          className={cn(
            "flex space-x-2 text-foreground/70 hover:text-foreground items-center min-w-0",
            className
          )}
        >
          <PackageIcon className="h-5 w-5 flex-none" />
          <span className="hidden lg:inline truncate max-w-48">
            {from.name}
          </span>
        </Link>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm">
        <div className="text-foreground/90">
          Added with the {from.name} {from.noun}. Updating it asks before
          replacing your changes. Click to open it.
        </div>
      </TooltipContent>
    </Tooltip>
  );
}
