import MessageNav from "@/components/message/MessageNav";
import { useBeforePageExit } from "@/lib/hooks/exit";
import MessageEditor from "@/tools/message-creator/components/MessageEditor";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useState } from "react";

export default function AppMessagePage() {
  const router = useRouter();

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const save = useCallback(() => {
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      setHasUnsavedChanges(false);
    }, 1000);
  }, []);

  const exit = useCallback(() => {
    if (hasUnsavedChanges) {
      if (
        !confirm("You have unsaved changes. Are you sure you want to exit?")
      ) {
        return;
      }
    }

    router.push({
      pathname: "/apps/[appId]/messages",
      query: { appId: router.query.appId },
    });
  }, [hasUnsavedChanges, router]);

  useBeforePageExit(
    (e) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        return "You have unsaved changes. Are you sure you want to exit?";
      }
    },
    [hasUnsavedChanges]
  );

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Message | Kite</title>
      </Head>
      <div className="h-[100dvh] w-[100dvw] flex flex-col">
        <div className="flex-none">
          <MessageNav
            hasUnsavedChanges={hasUnsavedChanges}
            isSaving={isSaving}
            onSave={save}
            onExit={exit}
          />
        </div>
        <div className="flex flex-auto overflow-y-hidden relative">
          <MessageEditor />
        </div>
      </div>
    </div>
  );
}
