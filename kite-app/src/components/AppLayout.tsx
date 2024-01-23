import clsx from "clsx";
import { Inter } from "next/font/google";
import { ReactNode } from "react";
import { Toaster } from "react-hot-toast";

const inter = Inter({ subsets: ["latin"] });

export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <div
      className={clsx(
        "min-h-[100dvh] bg-slate-700 pb-32 pt-10 px-5 lg:pt-32",
        inter.className
      )}
    >
      <div className="max-w-5xl mx-auto">{children}</div>
      <Toaster
        position="top-right"
        toastOptions={{
          className: "!bg-slate-800 !text-gray-100",
        }}
      />
    </div>
  );
}
