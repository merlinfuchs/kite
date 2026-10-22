import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import { MoonStarIcon, SunIcon } from "lucide-react";
import { Button } from "../ui/button";

export default function ThemeSwitch({ className }: { className?: string }) {
  // resolvedTheme is the theme that is shown, also when the setting is
  // "system". It's only known on the client, so show the moon until mounted.
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  const theme = mounted ? resolvedTheme : "dark";

  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn("-m-2", className)}
      aria-label="Toggle theme"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      {theme === "dark" ? (
        <MoonStarIcon className="h-6 w-6" />
      ) : (
        <SunIcon className="h-6 w-6" />
      )}
    </Button>
  );
}
