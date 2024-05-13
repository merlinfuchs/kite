import { ReactNode } from "react";
import { useUserQuery } from "@/lib/api/queries";
import LoginPrompt from "../LoginPrompt";
import BaseLayout from "@/components/BaseLayout";

interface Props {
  children: ReactNode;
  title?: string;
}

export default function AppsLayout({ children, title }: Props) {
  const { data: userResp } = useUserQuery();

  return (
    <BaseLayout title={title}>
      {!userResp ? null : !userResp.success ? <LoginPrompt /> : children}
    </BaseLayout>
  );
}
