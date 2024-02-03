import clsx from "clsx";
import { Inter } from "next/font/google";
import { ReactNode } from "react";
import { Toaster } from "react-hot-toast";
import Head from "next/head";

const inter = Inter({ subsets: ["latin"] });

interface Props {
  children: ReactNode;
  title?: string;
}

export default function HomeLayout({ children, title }: Props) {
  return (
    <div className={clsx("min-h-[100dvh]", inter.className)}>
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
      <Toaster
        position="top-right"
        toastOptions={{
          className: "!bg-dark-2 !text-gray-100",
        }}
      />
    </div>
  );
}
