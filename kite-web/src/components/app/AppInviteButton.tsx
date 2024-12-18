import Link from "next/link";
import { Button } from "../ui/button";
import { ChevronDownIcon, ExternalLinkIcon } from "lucide-react";
import { useApp } from "@/lib/hooks/api";
import { getAppInviteUrl } from "@/lib/discord/invite";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

export default function AppInviteButton() {
  const app = useApp();

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
        <DropdownMenuItem>
          <Link
            href={getAppInviteUrl(app.discord_id)}
            target="_blank"
            className="w-full flex items-center gap-2"
          >
            <ExternalLinkIcon className="h-3.5 w-3.5" />
            <span className="xl:whitespace-nowrap">Add app to server</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Link
            href={getAppInviteUrl(app.discord_id, "user")}
            target="_blank"
            className="w-full flex items-center gap-2"
          >
            <ExternalLinkIcon className="h-3.5 w-3.5" />
            <span className="xl:whitespace-nowrap">Add app to account</span>
          </Link>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
