import "@/styles/globals.css";
import "@/styles/shadow.css";
import "@/styles/message-preview.css";
import type { NextPage } from "next";
import type { AppProps } from "next/app";
import { useEffect, type ReactElement, type ReactNode } from "react";
import { Inter as FontSans } from "next/font/google";
import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider, useTheme } from "next-themes";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClientProvider } from "@tanstack/react-query";
import queryClient from "@/lib/api/client";
import AnalyticsProvider from "@/components/common/AnalyticsProvider";

function ServiceWorkerRegistration() {
  useEffect(() => {
    if (typeof window !== "undefined" && "serviceWorker" in navigator) {
      const register = () => {
        navigator.serviceWorker.register("/sw.js").catch(() => {});
      };
      if (document.readyState === "complete") {
        register();
      } else {
        window.addEventListener("load", register);
        return () => window.removeEventListener("load", register);
      }
    }
  }, []);

  return null;
}

function ThemeMetaSync() {
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    function syncTheme() {
      const isDark =
        document.documentElement.classList.contains("dark") ||
        resolvedTheme === "dark";
      const themeColor = isDark ? "#141211" : "#FCFCFB";

      let meta = document.querySelector('meta[name="theme-color"]');
      if (!meta) {
        meta = document.createElement("meta");
        meta.setAttribute("name", "theme-color");
        document.head.appendChild(meta);
      }
      meta.removeAttribute("media");
      meta.setAttribute("content", themeColor);

      let appleStatusBar = document.querySelector(
        'meta[name="apple-mobile-web-app-status-bar-style"]'
      );
      if (!appleStatusBar) {
        appleStatusBar = document.createElement("meta");
        appleStatusBar.setAttribute(
          "name",
          "apple-mobile-web-app-status-bar-style"
        );
        document.head.appendChild(appleStatusBar);
      }
      appleStatusBar.setAttribute(
        "content",
        isDark ? "black-translucent" : "default"
      );
    }

    syncTheme();

    const observer = new MutationObserver(() => syncTheme());
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });

    return () => observer.disconnect();
  }, [resolvedTheme]);

  return null;
}

const fontSans = FontSans({
  subsets: ["latin"],
  variable: "--font-sans",
});

type NextPageWithLayout = NextPage & {
  getLayout?: (page: ReactElement) => ReactNode;
};

type AppPropsWithLayout = AppProps & {
  Component: NextPageWithLayout;
};

export default function App({ Component, pageProps }: AppPropsWithLayout) {
  const getLayout = Component.getLayout ?? ((page) => page);

  return (
    <>
      <style jsx global>{`
        html {
          font-family: ${fontSans.style.fontFamily};
        }
      `}</style>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider attribute="class">
          <ThemeMetaSync />
          <ServiceWorkerRegistration />
          <TooltipProvider delayDuration={200}>
            {getLayout(<Component {...pageProps} />)}
            <Toaster position="top-right" richColors={true} />

            <AnalyticsProvider />
          </TooltipProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </>
  );
}
