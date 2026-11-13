import { UserBadges } from "@/components/badges/Badges";
import {
  ClipboardCheckIcon,
  HomeIcon,
  KeyRoundIcon,
  LogOutIcon,
  MailIcon,
} from "lucide-react";
import { ComponentProps, ReactNode, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/router";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useUser, useUserInvites } from "@/lib/hooks/api";
import { useLogout } from "@/lib/hooks/auth";
import { cn } from "@/lib/utils";
import APITokensDialog from "./APITokensDialog";
import UserInvitesDialog from "./UserInvitesDialog";

// The menu of the logged in user, opened by children. Pages that link home
// elsewhere, like with the logo, leave out returnHome.
//
// Guides that need a token link to ?api-tokens, which opens the API tokens.
// ?invites opens the invites to other apps.
export default function UserMenu({
  children,
  className,
  side,
  align = "end",
  returnHome = true,
}: {
  children: ReactNode;
  className?: string;
  side?: ComponentProps<typeof DropdownMenuContent>["side"];
  align?: ComponentProps<typeof DropdownMenuContent>["align"];
  returnHome?: boolean;
}) {
  const user = useUser();
  const logout = useLogout();
  const router = useRouter();
  const [tokensOpen, setTokensOpen] = useState(false);
  const tokensLinked = router.query["api-tokens"] !== undefined;
  const invites = useUserInvites();
  const [invitesOpen, setInvitesOpen] = useState(false);
  const invitesLinked = router.query["invites"] !== undefined;

  function onTokensOpenChange(open: boolean) {
    setTokensOpen(open);
    if (!open && tokensLinked) {
      const { "api-tokens": _, ...query } = router.query;
      router.replace({ pathname: router.pathname, query }, undefined, {
        shallow: true,
      });
    }
  }

  function onInvitesOpenChange(open: boolean) {
    setInvitesOpen(open);
    if (!open && invitesLinked) {
      const { invites: _, ...query } = router.query;
      router.replace({ pathname: router.pathname, query }, undefined, {
        shallow: true,
      });
    }
  }

  return (
    <>
      <APITokensDialog
        open={tokensOpen || tokensLinked}
        onOpenChange={onTokensOpenChange}
      />
      <UserInvitesDialog
        open={invitesOpen || invitesLinked}
        onOpenChange={onInvitesOpenChange}
      />
      <DropdownMenu>
        <DropdownMenuTrigger asChild>{children}</DropdownMenuTrigger>
        <DropdownMenuContent
          className={cn("min-w-56 rounded-lg", className)}
          side={side}
          align={align}
          sideOffset={4}
        >
          <DropdownMenuLabel className="p-0 font-normal">
            <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
              <UserAvatar className="h-8 w-8" />
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">
                  {user?.display_name}
                </span>
                <span className="truncate text-xs">
                  {user && `@${user.discord_username}`}
                </span>
              </div>
            </div>
            <UserBadges
              badges={user?.badges ?? []}
              size="sm"
              className="mx-1 mb-1.5"
            />
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          {returnHome && (
            <DropdownMenuItem className="gap-2 cursor-pointer" asChild>
              <Link href="/">
                <HomeIcon className="size-4" />
                Return home
              </Link>
            </DropdownMenuItem>
          )}
          {user?.is_admin && (
            <DropdownMenuItem className="gap-2 cursor-pointer" asChild>
              <Link href="/admin/reviews">
                <ClipboardCheckIcon className="size-4" />
                Marketplace reviews
              </Link>
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            className="gap-2 cursor-pointer"
            onClick={() => setInvitesOpen(true)}
          >
            <MailIcon className="size-4" />
            Invites
            {!!invites?.length && (
              <span className="ml-auto flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 text-xs font-semibold text-white">
                {invites.length}
              </span>
            )}
          </DropdownMenuItem>
          <DropdownMenuItem
            className="gap-2 cursor-pointer"
            onClick={() => setTokensOpen(true)}
          >
            <KeyRoundIcon className="size-4" />
            API tokens
          </DropdownMenuItem>
          <DropdownMenuItem className="gap-2 cursor-pointer" onClick={logout}>
            <LogOutIcon className="size-4" />
            Log out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}

// UserAvatar shows the initials of the logged in user.
export function UserAvatar({ className }: { className?: string }) {
  const user = useUser();
  const initials = (user?.display_name ?? "")
    .split(" ")
    .map((word) => word[0])
    .join("");

  return (
    <Avatar className={cn("rounded-lg", className)}>
      <AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
    </Avatar>
  );
}
