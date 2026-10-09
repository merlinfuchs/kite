import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

export function useHookedTheme() {
  const { theme, resolvedTheme, setTheme } = useTheme();

  const [realTheme, setRealTheme] = useState<string | undefined>("light");
  useEffect(() => {
    setRealTheme(resolvedTheme || theme);
  }, [theme, resolvedTheme]);

  return {
    theme: realTheme,
    setTheme,
  };
}
