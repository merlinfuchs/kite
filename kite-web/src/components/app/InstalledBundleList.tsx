import { Card } from "@/components/ui/card";
import { bundleInstallHref, hasUpdate } from "@/lib/flow/bundleInstalls";
import { templateIcons } from "@/lib/flow/templates";
import { useAppId } from "@/lib/hooks/params";
import { BundleInstall } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import { BadgeCheckIcon, PackageIcon } from "lucide-react";
import Link from "next/link";
import { Badge } from "../ui/badge";

// InstalledBundleList lists the templates and bundles the app added, the ones
// with an update first.
export default function InstalledBundleList({
  installs,
}: {
  installs: BundleInstall[];
}) {
  const sorted = [...installs].sort(
    (a, b) =>
      Number(hasUpdate(b)) - Number(hasUpdate(a)) ||
      a.name.localeCompare(b.name)
  );

  return (
    <div className="space-y-3">
      {sorted.map((install) => (
        <InstalledBundleEntry key={install.id} install={install} />
      ))}
    </div>
  );
}

function InstalledBundleEntry({ install }: { install: BundleInstall }) {
  const appId = useAppId();
  const Icon = templateIcons[install.icon] ?? PackageIcon;
  const gone = install.available_version === null;

  return (
    <Card className="p-4">
      <Link
        href={bundleInstallHref(install, appId)}
        className="flex items-start gap-3 group"
      >
        <div className="h-10 w-10 bg-primary/15 flex-none rounded-lg flex items-center justify-center">
          <Icon className="h-5 w-5 text-primary" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-medium group-hover:underline truncate">
              {install.name}
            </span>
            {hasUpdate(install) && (
              <Badge>Update to version {install.available_version}</Badge>
            )}
          </div>
          <div className="flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground mt-1">
            <Maker install={install} />
            <span>·</span>
            <span>Version {install.bundle_version}</span>
          </div>
          {gone && (
            <div className="text-xs text-destructive mt-1">
              {install.removed
                ? "Kite took this template down for breaking the rules. Check what it added."
                : "Its maker deleted it, so it doesn't get updates anymore."}
            </div>
          )}
        </div>
      </Link>
    </Card>
  );
}

function Maker({ install }: { install: BundleInstall }) {
  if (install.official) {
    return (
      <span className="flex items-center gap-1">
        <BadgeCheckIcon className="h-3.5 w-3.5 text-primary" />
        Made by Kite
      </span>
    );
  }
  if (install.yours) return <span>Made by you</span>;
  if (!install.owner) return <span>Made by a user</span>;
  return (
    <span className="flex items-center gap-1.5">
      Made by
      <img
        src={userAvatar({
          id: install.owner.discord_id,
          avatar: install.owner.discord_avatar,
        })}
        alt=""
        className="h-3.5 w-3.5 rounded-full"
      />
      {install.owner.display_name}
    </span>
  );
}
