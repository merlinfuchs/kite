import { ReactNode } from "react";
import BaseLayout from "@/components/BaseLayout";
import HomeNavbar from "@/components/home/HomeNavbar";
import HomeFooter from "@/components/home/HomeFooter";

interface Props {
  children: ReactNode;
  title?: string;
}

export default function AppLayout({ children, title }: Props) {
  return (
    <BaseLayout title={title}>
      {children}
      <HomeFooter />
    </BaseLayout>
  );
}
