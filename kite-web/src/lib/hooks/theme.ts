import { useTheme } from "next-themes";
import { useEffect, useState } from "react";

export function useHookedTheme() {
  const { theme, setTheme } = useTheme();

  const [realTheme, setRealTheme] = useState<string | undefined>("light");
  useEffect(() => {
    setRealTheme(theme);
  }, [theme]);

  return {
    theme: realTheme,
    setTheme,
  };
}
