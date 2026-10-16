import FlowPage from "@/components/flow/FlowPage";
import { useFunctionUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useFunction } from "@/lib/hooks/api";
import { useAppId, useFunctionId } from "@/lib/hooks/params";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useState } from "react";
import { toast } from "sonner";

export default function AppFunctionPage() {
  const router = useRouter();
  const fn = useFunction((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load function: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unknown_function") {
        router.push({
          pathname: "/apps/[appId]/functions",
          query: { appId: router.query.appId },
        });
      }
    }
  });

  const updateMutation = useFunctionUpdateMutation(useAppId(), useFunctionId());

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const onChange = useCallback(() => {
    setHasUnsavedChanges(true);
  }, [setHasUnsavedChanges]);

  const save = useCallback(
    (data: FlowData, options?: { onSuccess?: () => void }) => {
      setIsSaving(true);

      updateMutation.mutate(
        {
          flow_source: data,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                "Function saved! Flows that call it use the new version right away."
              );
              options?.onSuccess?.();
            } else {
              toast.error(
                `Failed to update function: ${res.error.message} (${res.error.code})`
              );
            }
          },
          onSettled() {
            setIsSaving(false);
            setHasUnsavedChanges(false);
          },
        }
      );
    },
    [setIsSaving, setHasUnsavedChanges, updateMutation]
  );

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const exit = useCallback(() => {
    leave({
      pathname: "/apps/[appId]/functions",
      query: { appId: router.query.appId },
    });
  }, [leave, router]);

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Function | Kite</title>
      </Head>
      {fn && (
        <FlowPage
          flowData={fn.flow_source}
          context="function"
          hasUnsavedChanges={hasUnsavedChanges}
          onChange={onChange}
          isSaving={isSaving}
          onSave={save}
          onExit={exit}
        />
      )}
    </div>
  );
}
