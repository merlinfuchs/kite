import { useEffect } from "react";
import { useRouter } from "next/router";
import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Templates",
  },
];

// Templates moved into the marketplace, old links go to its Official tab.
export default function AppTemplatesPage() {
  const router = useRouter();

  useEffect(() => {
    if (!router.isReady) return;
    router.replace({
      pathname: "/apps/[appId]/marketplace",
      query: { appId: router.query.appId, tab: "official" },
    });
  }, [router]);

  return null;
}

AppTemplatesPage.getLayout = getAppShellLayout;
