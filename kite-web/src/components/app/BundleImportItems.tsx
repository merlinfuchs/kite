import { useState } from "react";
import {
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import LoadingButton from "../common/LoadingButton";
import {
  TemplateSettingFields,
  useTemplateGuild,
} from "./TemplateSettingFields";
import { useBundleImportMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useAppSecrets } from "@/lib/hooks/api";
import { ExportedBundle } from "@/lib/flow/bundleExport";
import { missingSecretNames } from "@/lib/flow/secrets";
import { track } from "@/lib/analytics";
import { CircleAlertIcon } from "lucide-react";
import { toast } from "sonner";

// BundleImportItems is the second step of importing several items: what to
// include and the settings. It's loaded when it's needed, finding the secrets
// the blocks use needs all block definitions.
export default function ImportItems({
  content,
  bundle,
  onBack,
  onImported,
}: {
  content: unknown;
  bundle: ExportedBundle;
  onBack: () => void;
  onImported: () => void;
}) {
  const appId = useAppId();
  const secrets = useAppSecrets();
  const importMutation = useBundleImportMutation(appId);

  const [skipped, setSkipped] = useState<Set<string>>(new Set());
  const [values, setValues] = useState<Record<string, string>>({});
  const [guildId, setGuildId] = useTemplateGuild();

  const items = bundle.items.filter((i) => !skipped.has(i.key));
  // Message templates and variables are always added, the blocks use them.
  const keys = [...items, ...bundle.messages, ...bundle.variables].map(
    (i) => i.key
  );

  const missingSecrets = missingSecretNames(
    bundle.flows,
    secrets?.map((s) => s?.name) ?? []
  );

  async function onImport() {
    const res = await importMutation.mutateAsync({
      content: content as Record<string, unknown>,
      settings: Object.fromEntries(
        Object.entries(values).filter(([, value]) => value)
      ),
      items: keys,
    });
    if (!res.success) {
      toast.error(`Failed to import: ${res.error.message} (${res.error.code})`);
      return;
    }

    track("bundle_imported", { count: res.data.items.length });
    toast.success(
      res.data.items.length === 1
        ? "Imported 1 item!"
        : `Imported ${res.data.items.length} items!`
    );
    if (missingSecrets.length) {
      toast.warning(
        `Some blocks use secrets your app doesn't have: ${missingSecrets.join(
          ", "
        )}. Create them under Secrets.`
      );
    }
    onImported();
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import from {bundle.name}</DialogTitle>
        <DialogDescription>
          Copies of these are added to your app, so you can change them like
          anything you built yourself.
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5">
        {!!bundle.settings.length && (
          <TemplateSettingFields
            settings={bundle.settings}
            values={values}
            onChange={setValues}
            guildId={guildId}
            onGuildChange={setGuildId}
          />
        )}

        <div>
          <div className="text-sm font-medium mb-2.5">Include</div>
          <div className="space-y-2.5">
            {bundle.items.map((item) => (
              <label
                key={item.key}
                className="flex items-center gap-2.5 text-sm cursor-pointer"
              >
                <Checkbox
                  checked={!skipped.has(item.key)}
                  onCheckedChange={(checked) => {
                    const next = new Set(skipped);
                    if (checked) next.delete(item.key);
                    else next.add(item.key);
                    setSkipped(next);
                  }}
                />
                {item.label}
              </label>
            ))}
            {bundle.messages.map((message) => (
              <label
                key={message.key}
                className="flex items-center gap-2.5 text-sm"
              >
                <Checkbox checked disabled />
                {message.label} message
              </label>
            ))}
            {bundle.variables.map((variable) => (
              <label
                key={variable.key}
                className="flex items-center gap-2.5 text-sm"
              >
                <Checkbox checked disabled />
                {variable.label} stored variable
              </label>
            ))}
          </div>
        </div>

        {!!missingSecrets.length && (
          <div className="flex gap-2 text-sm text-amber-600 dark:text-amber-500">
            <CircleAlertIcon className="h-4 w-4 flex-none mt-0.5" />
            <span>
              Some blocks use secrets your app doesn&apos;t have:{" "}
              {missingSecrets.join(", ")}. Create them under Secrets after
              importing.
            </span>
          </div>
        )}
      </div>

      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={onBack}>
          Back
        </Button>
        {keys.length ? (
          <LoadingButton onClick={onImport} loading={importMutation.isPending}>
            Import {keys.length} {keys.length === 1 ? "item" : "items"}
          </LoadingButton>
        ) : (
          <Button disabled>Nothing selected</Button>
        )}
      </DialogFooter>
    </>
  );
}
