import { useHookedTheme } from "@/lib/hooks/theme";
import { cn } from "@/lib/utils";
import { MoonStarIcon, SunIcon } from "lucide-react";
import { Button } from "../ui/button";

export default function ThemeSwitch({ className }: { className?: string }) {
  const { theme, setTheme } = useHookedTheme();

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
