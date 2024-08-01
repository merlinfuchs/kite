import { useAppUpdateMutation } from "@/lib/api/mutations";
import { Button } from "../ui/button";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useRouter } from "next/router";
import { useApp } from "@/lib/hooks/api";

export default function AppSettingsDisable() {
  const app = useApp();

  const router = useRouter();
  const updateMutation = useAppUpdateMutation(useAppId());

  function toggleEnabled() {
    if (!app) return;

    updateMutation.mutate(
      {
        name: app.name,
        description: app.description,
        enabled: !app.enabled,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success(app.enabled ? "App disabled!" : "App enabled!");
          } else {
            toast.error(
              `Failed to update app: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Button
      variant="outline"
      className="space-x-2 flex items-center"
      onClick={toggleEnabled}
    >
      <div>{app?.enabled ? "Disable" : "Enable"} app</div>
    </Button>
  );
}
