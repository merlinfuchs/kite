import {
  LayoutPanelLeftIcon,
  LogInIcon,
  MoonStarIcon,
  PackageIcon,
  SunIcon,
} from "lucide-react";
import HomeNavbarMenu from "./HomeNavbarMenu";
import { useTheme } from "next-themes";
import { useAfterMounted } from "@/lib/hooks/mounted";
import { Button } from "../ui/button";
import Link from "next/link";
import { useUser } from "@/lib/hooks/api";

export default function HomeNavbar() {
  const { theme, setTheme } = useAfterMounted(useTheme(), {
    theme: "light",
    setTheme: () => {},
  });

  const user = useUser();

  return (
    <div className="border-b py-2 px-5 flex justify-between items-center">
      <HomeNavbarMenu />
      <div className="flex items-center space-x-5">
        {theme === "dark" ? (
          <MoonStarIcon
            className="w-6 h-6 cursor-pointer"
            onClick={() => setTheme("light")}
          />
        ) : (
          <SunIcon
            className="w-6 h-6 cursor-pointer"
            onClick={() => setTheme("dark")}
          />
        )}
        {user ? (
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
