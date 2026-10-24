import { ReactNode, useState } from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { ShareCodeInput, ShareCodePanel } from "./ShareCode";
import {
  TemplateSettingFields,
  useTemplateGuild,
} from "./TemplateSettingFields";
import {
  useBundleImportMutation,
  useShareCodeResolveMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useAppSecrets } from "@/lib/hooks/api";
import { ExportedBundle, parseExportedBundle } from "@/lib/flow/bundleExport";
import { getReferencedSecretNames } from "@/lib/flow/secrets";
import { track } from "@/lib/analytics";
import { BracesIcon, CircleAlertIcon, KeyRoundIcon } from "lucide-react";
import { toast } from "sonner";

export default function BundleImportDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <BundleImportForm onImported={() => setOpen(false)} />
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so its state resets every time it's opened.
function BundleImportForm({ onImported }: { onImported: () => void }) {
  const [useJson, setUseJson] = useState(false);
  const [json, setJson] = useState("");
  const [code, setCode] = useState("");
  const [bundle, setBundle] = useState<{
    content: unknown;
    parsed: ExportedBundle;
  } | null>(null);

  const resolveMutation = useShareCodeResolveMutation();

  function load(content: unknown) {
    const parsed = parseExportedBundle(content);
    if (!parsed) {
      toast.error("Invalid share code, make sure it was made with Export");
      return;
    }
    setBundle({ content, parsed });
  }

  function onContinue() {
    if (useJson) {
      let content;
      try {
        content = JSON.parse(json);
      } catch {}
      load(content);
      return;
    }

    if (!code) {
      toast.error("Enter a share code");
      return;
    }
    resolveMutation.mutate(code, {
      onSuccess: (res) => {
        if (!res.success) {
          toast.error(
            res.error.code === "unknown_share_code"
              ? "Unknown share code"
              : `Failed to resolve share code: ${res.error.message} (${res.error.code})`
          );
          return;
        }
        if (res.data.type !== "bundle") {
          toast.error(
            "This code is for a single command, event listener or message template, import it on its page"
          );
          return;
        }
        load(res.data.data);
      },
    });
  }

  if (bundle) {
    return (
      <ImportItems
        content={bundle.content}
        bundle={bundle.parsed}
        onBack={() => setBundle(null)}
        onImported={onImported}
      />
    );
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Import items</DialogTitle>
        <DialogDescription>
          {useJson
            ? "Paste the JSON that was exported from another app."
            : "Enter the share code that was exported from another app."}
        </DialogDescription>
      </DialogHeader>

      {useJson ? (
        <Textarea
          value={json}
          onChange={(e) => setJson(e.target.value)}
          placeholder='{"commands": ...}'
          minRows={8}
          maxRows={8}
          className="resize-none break-all font-mono md:text-xs"
          autoFocus
        />
      ) : (
        <ShareCodePanel>
          <ShareCodeInput
            value={code}
            onChange={setCode}
            onSubmit={() => !resolveMutation.isPending && onContinue()}
          />
        </ShareCodePanel>
      )}

      <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
        <Button variant="ghost" onClick={() => setUseJson(!useJson)}>
          {useJson ? (
            <KeyRoundIcon className="mr-2 h-4 w-4" />
          ) : (
            <BracesIcon className="mr-2 h-4 w-4" />
          )}
          {useJson ? "Use share code" : "Use JSON"}
        </Button>
        <div className="flex flex-col-reverse gap-2 sm:flex-row">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton
            onClick={onContinue}
            loading={resolveMutation.isPending}
          >
            Continue
          </LoadingButton>
        </div>
      </DialogFooter>
    </>
  );
}

function ImportItems({
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

  const secretNames = new Set(secrets?.map((s) => s?.name));
  const missingSecrets = Array.from(
    new Set(bundle.flows.flatMap((f) => getReferencedSecretNames(f)))
  ).filter((name) => !secretNames.has(name));

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
