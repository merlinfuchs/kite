import clsx from "clsx";
import { Inter } from "next/font/google";
import { ReactNode, useState } from "react";
import { Toaster } from "react-hot-toast";
import AppSidebar from "./AppSidebar";
import AppTopbar from "./AppTopbar";

const inter = Inter({ subsets: ["latin"] });

export default function AppLayout({ children }: { children: ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className={clsx("min-h-[100dvh]", inter.className)}>
      <AppTopbar setSidebarOpen={setSidebarOpen} />
      <div className="flex">
        <AppSidebar open={sidebarOpen} setOpen={setSidebarOpen} />
        <div className="pb-32 pt-10 xl:pt-20 px-5 md:px-10 flex flex-auto w-full justify-center">
          <div className="max-w-5xl w-full">{children}</div>
        </div>
      </div>
      <Toaster
        position="top-right"
        toastOptions={{
          className: "!bg-dark-2 !text-gray-100",
        }}
      />
    </div>
  );
}
