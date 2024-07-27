import { useHookedTheme } from "@/lib/hooks/theme";
import { MoonStarIcon, SunIcon } from "lucide-react";
import { useEffect, useState } from "react";

export default function ThemeSwitch() {
  const { theme, setTheme } = useHookedTheme();

  return (
    <>
      {theme === "dark" ? (
        <MoonStarIcon
          className="h-6 w-6 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
          onClick={() => setTheme("light")}
        />
      ) : (
        <SunIcon
          className="h-6 w-6 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
          onClick={() => setTheme("dark")}
        />
      )}
    </>
  );
}
