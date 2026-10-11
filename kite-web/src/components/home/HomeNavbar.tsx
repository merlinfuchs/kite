import { LayoutPanelLeftIcon, LogInIcon, PackageIcon } from "lucide-react";
import HomeNavbarMenu from "./HomeNavbarMenu";
import { Button } from "../ui/button";
import Link from "next/link";
import { useResponseData } from "@/lib/hooks/api";
import { useUserQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import ThemeSwitch from "../common/ThemeSwitch";
import { loginUrl } from "@/lib/api/client";

export default function HomeNavbar() {
  const userQuery = useUserQuery();
  const user = useResponseData(userQuery);

  return (
    <div className="border-b py-2 px-5 flex justify-between items-center">
      <HomeNavbarMenu />
      <div className="flex items-center space-x-5">
        <ThemeSwitch />
        {userQuery.isPending ? (
          <Skeleton className="h-10 w-28" />
        ) : user ? (
          <Button asChild>
            <Link href="/apps" className="flex items-center space-x-1.5">
              <PackageIcon className="h-5 w-5" />
              <div>Open app</div>
            </Link>
          </Button>
        ) : (
          <Button asChild>
            <a href={loginUrl} className="flex items-center space-x-2">
              <LogInIcon className="h-5 w-5" />
              <div>Login</div>
            </a>
          </Button>
        )}
      </div>
    </div>
  );
}
