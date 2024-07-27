import Flow from "@/components/flow/Flow";
import { useCommandUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/data";
import { useCommand } from "@/lib/hooks/api";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import Head from "next/head";
import { useRouter } from "next/router";
import { useState } from "react";
import { toast } from "sonner";

export default function AppCommandPage() {
  const router = useRouter();
  const cmd = useCommand((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load command: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unknown_command") {
        router.push({
          pathname: "/apps/[appId]/commands",
          query: { appId: router.query.appId },
        });
      }
    }
  });

  const updateMutation = useCommandUpdateMutation(useAppId(), useCommandId());

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  function save(data: FlowData) {
    setIsSaving(true);

    updateMutation.mutate(
      {
        flow_source: data as any, // TODO
        enabled: true,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success(
              "Command saved! Is may take up to a minute for all changes to take effect."
            );
          } else {
            toast.error(
              `Failed to update command: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onSettled() {
          setIsSaving(false);
          setHasUnsavedChanges(false);
        },
      }
    );
  }

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Command | Kite</title>
      </Head>
      {cmd && (
        <Flow
          flowData={cmd.flow_source as any} // TODO
          hasUnsavedChanges={hasUnsavedChanges}
          onChange={() => setHasUnsavedChanges(true)}
          isSaving={isSaving}
          onSave={save}
          onExit={() =>
            router.push({
              pathname: "/apps/[appId]/commands",
              query: { appId: router.query.appId },
            })
          }
        />
      )}
    </div>
  );
}
