import { LogOutIcon } from "lucide-react";
import BaseUserAvatar from "@/components/common/UserAvatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { User } from "@/lib/types/wire.gen";
import { useLogout } from "@/lib/hooks/auth";

export function UserAvatar({
  user,
  className,
}: {
  user: User;
  className?: string;
}) {
  return (
    <BaseUserAvatar displayName={user.display_name} className={className} />
  );
}

export function UserInfo({ user }: { user: User }) {
  return (
    <div className="grid flex-1 text-left text-sm leading-tight">
      <span className="truncate font-semibold">{user.display_name}</span>
      <span className="truncate text-xs">{user.email}</span>
    </div>
  );
}

export default function HomeNavbarUser({ user }: { user: User }) {
  const logout = useLogout();

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger
        className="rounded-full outline-none ring-offset-background transition-shadow hover:ring-2 hover:ring-primary/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        aria-label="Account menu"
      >
        <UserAvatar user={user} />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        sideOffset={8}
        className="min-w-56 rounded-lg"
      >
        <DropdownMenuLabel className="p-0 font-normal">
          <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
            <UserAvatar user={user} />
            <UserInfo user={user} />
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="gap-2 cursor-pointer" onClick={logout}>
          <LogOutIcon className="size-4" />
          Log out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
