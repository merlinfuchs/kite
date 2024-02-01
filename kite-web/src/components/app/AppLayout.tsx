import clsx from "clsx";
import { Inter } from "next/font/google";
import { ReactNode } from "react";
import { Toaster } from "react-hot-toast";
import Head from "next/head";
import { useUserQuery } from "@/lib/api/queries";
import LoginPrompt from "../LoginPrompt";

const inter = Inter({ subsets: ["latin"] });

export default function AppLayout({ children }: { children: ReactNode }) {
  const { data: userResp } = useUserQuery();

  return (
    <div className={clsx("min-h-[100dvh]", inter.className)}>
      <Head>
        <title>Kite.onl | Discord Bots made easy</title>
      </Head>
      {!userResp ? null : !userResp.success ? <LoginPrompt /> : children}
      <Toaster
        position="top-right"
        toastOptions={{
          className: "!bg-dark-2 !text-gray-100",
        }}
      />
    </div>
  );
}
