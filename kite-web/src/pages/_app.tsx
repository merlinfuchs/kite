import "@/styles/globals.css";
import "@/styles/shadow.css";
import "@/styles/message-preview.css";
import type { NextPage } from "next";
import type { AppProps } from "next/app";
import type { ReactElement, ReactNode } from "react";
import { Inter as FontSans } from "next/font/google";
import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider } from "next-themes";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClientProvider } from "@tanstack/react-query";
import queryClient from "@/lib/api/client";
import AnalyticsProvider from "@/components/common/AnalyticsProvider";

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
