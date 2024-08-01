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
      <div className="min-h-[100dvh] flex flex-col overflow-hidden">
        <div className="flex-none">
          <HomeNavbar />
        </div>
        <div className="flex-auto overflow-hidden">{children}</div>
      </div>
    </BaseLayout>
  );
}
