import { ReactNode, useState } from "react";
import AppsLayout from "./AppsLayout";
import AppSidebar from "./AppSidebar";
import AppTopbar from "./AppTopbar";

export default function AppLayout({ children }: { children: ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <AppsLayout>
      <AppTopbar setSidebarOpen={setSidebarOpen} />
      <div className="flex">
        <AppSidebar open={sidebarOpen} setOpen={setSidebarOpen} />
        <div className="pb-32 pt-10 xl:pt-20 px-5 md:px-10 flex flex-auto w-full justify-center">
          <div className="max-w-5xl w-full">{children}</div>
        </div>
      </div>
    </AppsLayout>
  );
}
