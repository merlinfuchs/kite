import { ReactNode } from "react";
import HomeNavbar from "./HomeNavbar";
import Head from "next/head";
import BaseLayout from "../common/BaseLayout";

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
    <BaseLayout title={title} description={description}>
      <div className="min-h-[100dvh] flex flex-col overflow-clip">
        <div className="flex-none sticky top-0 z-50">
          <HomeNavbar />
        </div>
        <div className="flex-auto overflow-clip">{children}</div>
      </div>
    </BaseLayout>
  );
}
