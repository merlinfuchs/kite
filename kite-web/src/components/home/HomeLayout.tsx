import { ReactNode } from "react";
import HomeNavbar from "./HomeNavbar";
import BaseLayout from "../common/BaseLayout";

export default function HomeLayout({
  children,
  ...props
}: {
  children: ReactNode;
  title?: string;
  description?: string;
  noindex?: boolean;
  jsonLd?: object[];
}) {
  return (
    <BaseLayout {...props}>
      {/* overflow-x-clip instead of overflow-hidden, so the sticky navbar works */}
      <div className="min-h-[100dvh] flex flex-col overflow-x-clip">
        <HomeNavbar />
        <div className="flex-auto overflow-hidden">{children}</div>
      </div>
    </BaseLayout>
  );
}
