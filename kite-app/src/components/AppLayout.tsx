import clsx from "clsx";
import { Inter } from "next/font/google";
import { ReactNode, useState } from "react";
import { Toaster } from "react-hot-toast";
import AppSidebar from "./AppSidebar";
import AppTopbar from "./AppTopbar";
import Head from "next/head";

const inter = Inter({ subsets: ["latin"] });

export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <div className={clsx("min-h-[100dvh]", inter.className)}>
      <Head>
        <title>Kite.onl | Discord Bots made easy</title>
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
