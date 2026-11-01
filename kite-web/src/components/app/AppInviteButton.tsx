import Link from "next/link";
import { Button } from "../ui/button";
import { ChevronDownIcon, ExternalLinkIcon } from "lucide-react";
import { useApp } from "@/lib/hooks/api";
import { getAppInviteUrl } from "@/lib/discord/invite";
import { useAppInstallationQuery } from "@/lib/api/queries";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

export default function AppInviteButton() {
  const app = useApp();

  // The links use the installation settings of the app. Until they are
  // loaded, or if they can't be, both links work without them.
  const installationQuery = useAppInstallationQuery(app?.id ?? "");
  const installation = installationQuery.data?.success
    ? installationQuery.data.data
    : undefined;

  if (!app) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="sm" variant="default" className="h-8 gap-2">
          <ChevronDownIcon className="h-3.5 w-3.5" />

          <span className="xl:whitespace-nowrap">Invite app</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <InviteItem
          label="Add app to server"
          disabledReason="Server install is turned off in the installation settings of your app."
          url={getAppInviteUrl(
            app.discord_id,
            "guild",
            installation?.guild_install
          )}
          enabled={!installation || !!installation.guild_install}
        />
        <InviteItem
          label="Add app to account"
          disabledReason="Account install is turned off in the installation settings of your app."
          url={getAppInviteUrl(
            app.discord_id,
            "user",
            installation?.user_install
          )}
          enabled={!installation || !!installation.user_install}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function InviteItem({
  label,
  disabledReason,
  url,
  enabled,
}: {
  label: string;
  disabledReason: string;
  url: string;
  enabled: boolean;
}) {
  if (!enabled) {
    // Disabled items ignore the pointer, so the tooltip is on a wrapper.
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="cursor-not-allowed">
            <DropdownMenuItem disabled className="flex items-center gap-2">
              <ExternalLinkIcon className="h-3.5 w-3.5" />
              <span className="xl:whitespace-nowrap">{label}</span>
            </DropdownMenuItem>
          </div>
        </TooltipTrigger>
        <TooltipContent side="left" className="max-w-56">
          {disabledReason}
        </TooltipContent>
      </Tooltip>
    );
  }

  return (
    <DropdownMenuItem>
      <Link
        href={url}
        target="_blank"
        className="w-full flex items-center gap-2"
      >
        <ExternalLinkIcon className="h-3.5 w-3.5" />
        <span className="xl:whitespace-nowrap">{label}</span>
      </Link>
    </DropdownMenuItem>
  );
}
