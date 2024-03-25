import { ReactNode } from "react";
import { Toaster } from "@/components/ui/sonner";
import Head from "next/head";

interface Props {
  children: ReactNode;
  title?: string;
}

export default function BaseLayout({ children, title }: Props) {
  return (
    <div className="min-h-[100dvh]">
      <Head>
        <title>{"Kite.onl | " + (title || "Discord Bots made easy")}</title>
        <meta
          name="description"
          content="Make Discord Bots without worrying about hosting and scaling. Concentrate on what you do best, building your bot."
        />
        <meta
          name="og:description"
          content="Make Discord Bots without worrying about hosting and scaling. Concentrate on what you do best, building your bot."
        />
        <meta name="og:title" content="Kite.onl | Discord Bots made easy" />
        <meta name="og:site_name" content="kite.onl" />
      </Head>
      {children}
      <Toaster position="top-right" />
    </div>
  );
}
