import { Button } from "@/components/ui/button";
import { useUser } from "@/lib/hooks/api";
import { ChevronDownIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";
import logo from "@/assets/logo/orange@1024.png";
import UserMenu, { UserAvatar } from "./UserMenu";

// AccountPageLayout is the layout of pages outside of an app, like the app
// list, with the logo and the account menu above a centered column.
export default function AccountPageLayout({
  children,
  showAccount = true,
}: {
  children: ReactNode;
  showAccount?: boolean;
}) {
  const user = useUser();

  return (
    <div className="flex min-h-[100dvh] flex-col">
      <header className="flex h-14 flex-none items-center justify-between border-b px-5">
        <Link
          href="/"
          className="flex items-center gap-2 rounded-md font-semibold outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <img src={logo.src} className="h-6 w-6" alt="" />
          Kite
        </Link>
        {showAccount && (
          <UserMenu returnHome={false}>
            <Button
              variant="ghost"
              size="sm"
              className="gap-2"
              aria-label="Account menu"
            >
              <UserAvatar className="size-6 text-xs" />
              <span className="hidden sm:inline">{user?.display_name}</span>
              <ChevronDownIcon className="size-4 text-muted-foreground" />
            </Button>
          </UserMenu>
        )}
      </header>
      <div className="flex flex-1 justify-center items-center w-full px-5 pt-10 pb-20">
        <div className="w-full max-w-lg">{children}</div>
      </div>
    </div>
  );
}
