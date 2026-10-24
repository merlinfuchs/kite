import { Template } from "@/lib/flow/templates";
import { eventListenerLabel } from "@/lib/flow/eventTypes";
import {
  useBundleInstallDeleteMutation,
  useBundleInstallParametersMutation,
  useBundleInstallUpdateMutation,
  useMessageInstanceCreateMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { useBundleInstalls } from "@/lib/hooks/api";
import { BundleInstall } from "@/lib/types/wire.gen";
import { track } from "@/lib/analytics";
import { ArrowUpCircleIcon, CircleAlertIcon } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import { Checkbox } from "../ui/checkbox";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import ConfirmDialog from "../common/ConfirmDialog";
import LoadingButton from "../common/LoadingButton";
import {
  TemplateParameterFields,
  useTemplateGuild,
} from "./TemplateParameterFields";

// TemplateInstalledCard shows that the app has the template, with updating it
// to a new version and removing it.
export function TemplateInstalledCard({ template }: { template: Template }) {
  const appId = useAppId();
  // An app has a template once.
  const install = useBundleInstalls()?.find(
    (i): i is BundleInstall => i?.bundle_id === template.id
  );
  const updateMutation = useBundleInstallUpdateMutation(appId);
  const deleteMutation = useBundleInstallDeleteMutation(appId);

  const [updateOpen, setUpdateOpen] = useState(false);
  const [overwrite, setOverwrite] = useState<Set<string>>(new Set());
  const [values, setValues] = useState<Record<string, string>>({});
  const [guildId, setGuildId] = useTemplateGuild();

  const items = install?.items.filter((item) => !item.deleted) ?? [];
  if (!install || !items.length) return null;

  const outdated = install.bundle_version < template.version;
  // Items that are no longer in the template stay as they are, and so do
  // variables, which only have a name.
  const updatable = outdated
    ? items.filter(
        (item) => item.type !== "variable" && templateItem(template, item.key)
      )
    : [];
  const edited = updatable.filter((item) => item.edited);
  const unedited = updatable.filter((item) => !item.edited);
  const replaced = updatable.filter(
    (item) => !item.edited || overwrite.has(item.key)
  );

  // Parameters the new version uses that the install has no value for.
  const usedKeys = new Set(
    replaced.flatMap((item) => templateItem(template, item.key)!.parameterKeys)
  );
  const missingParameters = template.parameters.filter(
    (s) => usedKeys.has(s.key) && !install.parameters?.[s.key]
  );

  async function update() {
    const res = await updateMutation.mutateAsync({
      installId: install!.id,
      bundle_version: template.version,
      parameters: Object.fromEntries(
        missingParameters.flatMap((s) =>
          values[s.key] ? [[s.key, values[s.key]]] : []
        )
      ),
      overwrite: Array.from(overwrite),
    });
    if (!res.success) {
      toast.error(
        `Failed to update ${template.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    track("template_updated", {
      template: template.id,
      template_version: template.version,
    });
    setOverwrite(new Set());
    setUpdateOpen(false);
    toast.success(`Updated ${template.name} to version ${template.version}`);
  }

  async function remove() {
    const res = await deleteMutation.mutateAsync(install!.id);
    if (!res.success) {
      toast.error(
        `Failed to remove ${template.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    track("template_removed", { template: template.id });
    toast.success(`Removed ${template.name} from your app`);
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">In your app</div>
      <p className="text-sm text-muted-foreground mb-4">
        {items.length === 1
          ? "1 item of this template is in your app."
          : `${items.length} items of this template are in your app.`}
      </p>

      <InstallParameters template={template} install={install} />

      {outdated && (
        <div className="mb-5">
          <div className="flex items-center gap-2 text-sm font-medium mb-1.5">
            <ArrowUpCircleIcon className="h-4 w-4 text-primary" />
            Version {template.version} is available
          </div>
          <p className="text-sm text-muted-foreground mb-3">
            Updating replaces what you didn&apos;t change with the new version.
          </p>
          <Button className="w-full" onClick={() => setUpdateOpen(true)}>
            Update to version {template.version}
          </Button>
        </div>
      )}

      <Dialog open={updateOpen} onOpenChange={setUpdateOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              Update {template.name} to version {template.version}
            </DialogTitle>
            <DialogDescription>
              Messages already posted in Discord stay as they are.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-5">
            {!!unedited.length && (
              <div>
                <div className="text-sm font-medium mb-1.5">
                  Replaced with the new version
                </div>
                <div className="text-sm text-muted-foreground">
                  {new Intl.ListFormat("en").format(
                    unedited.map((item) => itemLabel(template, item.key))
                  )}
                </div>
              </div>
            )}

            {!!edited.length && (
              <div>
                <div className="text-sm font-medium mb-1.5">
                  You changed these
                </div>
                <div className="text-sm text-muted-foreground mb-3">
                  They keep your changes unless you choose to replace them.
                </div>
                <div className="space-y-2.5">
                  {edited.map((item) => (
                    <label
                      key={item.id}
                      className="flex items-center gap-2.5 text-sm cursor-pointer"
                    >
                      <Checkbox
                        checked={overwrite.has(item.key)}
                        onCheckedChange={(checked) => {
                          const next = new Set(overwrite);
                          if (checked) next.add(item.key);
                          else next.delete(item.key);
                          setOverwrite(next);
                        }}
                      />
                      <span>{itemLabel(template, item.key)}</span>
                      <span className="text-xs text-muted-foreground ml-auto">
                        replace my changes
                      </span>
                    </label>
                  ))}
                </div>
              </div>
            )}

            {!!missingParameters.length && (
              <div>
                <div className="text-sm font-medium mb-3">New parameters</div>
                <TemplateParameterFields
                  parameters={missingParameters}
                  values={values}
                  onChange={setValues}
                  guildId={guildId}
                  onGuildChange={setGuildId}
                />
              </div>
            )}
          </div>

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <LoadingButton onClick={update} loading={updateMutation.isPending}>
              Update
            </LoadingButton>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <PostMessages template={template} install={install} />

      <ConfirmDialog
        title={`Remove ${template.name}?`}
        description={`This deletes ${new Intl.ListFormat("en").format(
          items.map((item) => itemLabel(template, item.key))
        )} from your app, including your changes to them${
          items.some((item) => item.type === "variable")
            ? " and the values stored in the variables"
            : ""
        }. Message templates and variables that other commands or event listeners use are kept.`}
        onConfirm={remove}
      >
        <Button
          variant="outline"
          className="w-full"
          disabled={deleteMutation.isPending}
        >
          Remove from app
        </Button>
      </ConfirmDialog>
    </Card>
  );
}

// PostMessages posts the messages that the template posted when it was added
// once more, in channels the user picks. Posted ones keep working until
// they're deleted in Discord.
function PostMessages({
  template,
  install,
}: {
  template: Template;
  install: BundleInstall;
}) {
  const appId = useAppId();
  const postMutation = useMessageInstanceCreateMutation(appId);
  const [open, setOpen] = useState(false);
  const [posting, setPosting] = useState(false);
  const [guildId, setGuildId] = useTemplateGuild();
  // Not the channels it was posted in, the install doesn't know their
  // server.
  const [values, setValues] = useState<Record<string, string>>({});

  const messages = template.messages.flatMap((message) => {
    const item = install.items.find(
      (i) => i.type === "message" && i.key === message.key && !i.deleted
    );
    return message.sendTo && item
      ? [{ name: message.name, sendTo: message.sendTo, id: item.id }]
      : [];
  });
  if (!messages.length) return null;
  const channels = template.parameters.filter((p) =>
    messages.some((m) => m.sendTo === p.key)
  );
  const label =
    messages.length === 1 ? `${messages[0]!.name} message` : "messages";

  async function post() {
    if (!guildId) return;
    setPosting(true);
    let posted = 0;
    try {
      for (const message of messages) {
        const res = await postMutation.mutateAsync({
          messageId: message.id,
          discord_guild_id: guildId,
          discord_channel_id: values[message.sendTo]!,
        });
        if (!res.success) {
          toast.error(
            `Failed to post the ${message.name} message: ${res.error.message} (${res.error.code})`
          );
        } else {
          track("message_sent", { source: "template", template: template.id });
          posted++;
        }
      }
    } catch (e) {
      toast.error(`Failed to post the ${label}: ${e}`);
    } finally {
      setPosting(false);
    }
    if (posted === messages.length) {
      toast.success(`Posted the ${label}`);
      setOpen(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button
        variant="outline"
        className="w-full mb-2"
        onClick={() => setOpen(true)}
      >
        Post the {label} again
      </Button>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Post the {label} again</DialogTitle>
          <DialogDescription>
            {messages.length === 1
              ? "Posts it in the channel you pick, like when the template was added. The one already posted keeps working until you delete it in Discord."
              : "Posts them in the channels you pick, like when the template was added. The ones already posted keep working until you delete them in Discord."}
          </DialogDescription>
        </DialogHeader>
        <TemplateParameterFields
          parameters={channels}
          values={values}
          onChange={setValues}
          guildId={guildId}
          onGuildChange={setGuildId}
        />
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton
            onClick={post}
            loading={posting}
            disabled={!guildId || messages.some((m) => !values[m.sendTo])}
          >
            Post
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// InstallParameters changes the parameters of the template in every flow it
// added. It shows the values picked when the template was added, and where a
// flow was changed in the editor since.
function InstallParameters({
  template,
  install,
}: {
  template: Template;
  install: BundleInstall;
}) {
  const appId = useAppId();
  const mutation = useBundleInstallParametersMutation(appId);
  const [guildId, setGuildId] = useTemplateGuild();
  const [values, setValues] = useState<Record<string, string> | null>(null);

  const items = install.items.filter((it) => !it.deleted);
  // The parameters of the template that the flows still use. Ones only used
  // to post a message, or added in the editor, aren't shown.
  const flowKeys = new Set(
    items.flatMap((it) => Object.keys(it.parameters ?? {}))
  );
  const fields = template.parameters.filter((p) => flowKeys.has(p.key));
  if (!fields.length) return null;
  const keys = fields.map((p) => p.key);

  // The values picked when the template was added.
  const saved = Object.fromEntries(
    keys.map((key) => [key, install.parameters?.[key] ?? ""])
  );
  const current = values ?? saved;
  // Only values changed here are saved, so the ones changed in the editor
  // stay unless they're changed here too.
  const changed = keys.filter((key) => (current[key] ?? "") !== saved[key]);

  const notes = Object.fromEntries(
    keys.flatMap((key) => {
      const changedIn = items.filter(
        (it) =>
          it.parameters?.[key] !== undefined &&
          it.parameters[key] !== saved[key]
      );
      if (!changedIn.length) return [];
      return [
        [
          key,
          <div
            key={key}
            className="flex gap-1.5 text-xs text-amber-600 dark:text-amber-500"
          >
            <CircleAlertIcon className="h-3.5 w-3.5 flex-none mt-px" />
            <span>
              Changed in{" "}
              {new Intl.ListFormat("en").format(
                changedIn.map((it) => itemLabel(template, it.key))
              )}
              . Saving a new value replaces it there too.
            </span>
          </div>,
        ],
      ];
    })
  );

  async function save() {
    const res = await mutation.mutateAsync({
      installId: install.id,
      parameters: Object.fromEntries(
        changed.map((key) => [key, current[key] ?? ""])
      ),
    });
    if (!res.success) {
      toast.error(
        `Failed to save the parameters: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    setValues(null);
    toast.success(`Saved the parameters of ${template.name}`);
  }

  return (
    <div className="mb-5">
      <div className="text-sm font-medium mb-1.5">Parameters</div>
      <p className="text-sm text-muted-foreground mb-3">
        Changes go into every command, event listener and message the template
        added, also into messages that were already posted.
      </p>
      <TemplateParameterFields
        parameters={fields}
        values={current}
        onChange={setValues}
        guildId={guildId}
        onGuildChange={setGuildId}
        notes={notes}
      />
      <LoadingButton
        className="w-full mt-4"
        variant="outline"
        onClick={save}
        loading={mutation.isPending}
        disabled={!changed.length}
      >
        Save parameters
      </LoadingButton>
    </div>
  );
}

function templateItem(template: Template, key: string) {
  const command = template.commands.find((c) => c.key === key);
  if (command) return { label: `/${command.name}`, ...command };
  const listener = template.eventListeners.find((l) => l.key === key);
  if (listener) {
    return {
      label: eventListenerLabel(listener.source, listener.type),
      ...listener,
    };
  }
  const variable = template.variables.find((v) => v.key === key);
  if (variable) {
    return {
      label: `${variable.name} stored variable`,
      ...variable,
      parameterKeys: [],
    };
  }
  const message = template.messages.find((m) => m.key === key);
  if (message) {
    return {
      label: `${message.name} message`,
      ...message,
    };
  }
}

function itemLabel(template: Template, key: string) {
  return templateItem(template, key)?.label ?? key;
}
