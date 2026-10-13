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
      <div className="min-h-[100dvh] flex flex-col overflow-hidden">
        <div className="flex-none">
          <HomeNavbar />
        </div>
        <div className="flex-auto overflow-hidden">{children}</div>
      </div>
    </BaseLayout>
  );
}
