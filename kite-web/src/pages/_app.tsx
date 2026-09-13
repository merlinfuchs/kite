import "@/styles/globals.css";
import "@/styles/shadow.css";
import type { AppProps } from "next/app";
import { Inter as FontSans } from "next/font/google";
import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider } from "next-themes";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClientProvider } from "@tanstack/react-query";
import queryClient from "@/lib/api/client";
import { SpeedInsights } from "@vercel/speed-insights/next";
import AnalyticsProvider from "@/components/common/AnalyticsProvider";
import { useEffect } from "react";

const fontSans = FontSans({
  subsets: ["latin"],
  variable: "--font-sans",
});

export default function App({ Component, pageProps }: AppProps) {
  useEffect(() => {
    // When a page is restored from the browser's back/forward cache
    // (e.g. clicking the browser Back button rather than an in-app link),
    // the browser reuses a frozen snapshot of the page instead of
    // re-running any JavaScript -- no component re-mounts, no effects
    // re-fire, no requests go out. Whatever React Query had cached (even
    // mid-load, or before an update landed) is what gets shown, with no
    // way for our own code to react to it. `pageshow`'s `persisted` flag
    // is how the browser tells us this just happened, so we force every
    // query to refetch now that the page is actually interactive again.
    const handlePageShow = (event: PageTransitionEvent) => {
      if (event.persisted) {
        queryClient.invalidateQueries();
      }
    };

    window.addEventListener("pageshow", handlePageShow);
    return () => window.removeEventListener("pageshow", handlePageShow);
  }, []);

  return (
    <>
      <style jsx global>{`
        html {
          font-family: ${fontSans.style.fontFamily};
        }
      `}</style>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider attribute="class">
          <TooltipProvider delayDuration={200}>
            <Component {...pageProps} />
            <Toaster position="top-right" richColors={true} />

            <SpeedInsights />
            <AnalyticsProvider />
          </TooltipProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </>
  );
}
