import { ReactNode, useState } from "react";
import AppLayout from "./AppLayout";
import AppSidebar from "./AppSidebar";
import AppTopbar from "./AppTopbar";

export default function AppGuildLayout({ children }: { children: ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <AppLayout>
      <AppTopbar setSidebarOpen={setSidebarOpen} />
      <div className="flex">
        <AppSidebar open={sidebarOpen} setOpen={setSidebarOpen} />
        <div className="pb-32 pt-10 xl:pt-20 px-5 md:px-10 flex flex-auto w-full justify-center">
          <div className="max-w-5xl w-full">{children}</div>
        </div>
      </div>
    </AppLayout>
  );
}
