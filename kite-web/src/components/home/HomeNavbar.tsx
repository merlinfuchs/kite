import { LayoutPanelLeftIcon, LogInIcon, PackageIcon } from "lucide-react";
import HomeNavbarMenu from "./HomeNavbarMenu";
import { Button } from "../ui/button";
import Link from "next/link";
import { useResponseData } from "@/lib/hooks/api";
import { useUserQuery } from "@/lib/api/queries";
import { Skeleton } from "../ui/skeleton";
import ThemeSwitch from "../common/ThemeSwitch";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

export default function HomeNavbar() {
  const userQuery = useUserQuery();
  const user = useResponseData(userQuery);

  // The navbar sticks to the top, and gets a blurred backdrop once the page
  // is scrolled so the content passing below stays readable.
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <div
      className={cn(
        "border-b py-2 px-5 flex justify-between items-center transition-[background-color,box-shadow,backdrop-filter] duration-300",
        scrolled
          ? "bg-background/75 backdrop-blur-md shadow-sm"
          : "bg-background"
      )}
    >
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
            <Link href="/login" className="flex items-center space-x-2">
              <LogInIcon className="h-5 w-5" />
              <div>Login</div>
            </Link>
          </Button>
        )}
      </div>
    </div>
  );
}
