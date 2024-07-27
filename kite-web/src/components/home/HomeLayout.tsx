import { ReactNode } from "react";
import HomeNavbar from "./HomeNavbar";
import Head from "next/head";

export default function HomeLayout({
  children,
  title,
  description,
}: {
  children: ReactNode;
  title?: string;
  description?: string;
}) {
  return (
    <div className="min-h-[100dvh] flex flex-col overflow-hidden">
      <Head>
        <title>{`${title ? title + " | " : ""}Kite`}</title>
        <meta
          name="description"
          content={
            description ||
            "Kite - Create Discord Bots for free with no coding required."
          }
        />
      </Head>
      <div className="flex-none">
        <HomeNavbar />
      </div>
      <div className="flex-auto overflow-hidden">{children}</div>
    </div>
  );
}
