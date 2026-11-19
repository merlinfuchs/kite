import { useMessageDataUpdateMutation } from "@/lib/api/mutations";
import { useMessageQuery } from "@/lib/api/queries";
import { Template } from "@/lib/flow/templates";
import { useApp, useBundleInstalls, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { ArrowLeftIcon } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";
import MessageEditorDialog from "../message/MessageEditorDialog";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import AppCreateDialog from "./AppCreateDialog";
import TemplateIcon from "./TemplateIcon";
import { TemplateInstallCard } from "./TemplateInstallCard";
import { TemplateInstalledCard } from "./TemplateInstalledCard";

// KiteBotTemplate is the page of a template for an app on the Kite bot: how
// it works and its settings, without the blocks inside like for custom bots.
export default function KiteBotTemplate({ template }: { template: Template }) {
  const appId = useAppId();
  const installs = useBundleInstalls();
  const added = installs?.some((i) => i?.bundle_id === template.id);

  return (
    <div className="space-y-6 max-w-3xl">
      <Link
        href={{ pathname: "/apps/[appId]", query: { appId } }}
        className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeftIcon className="h-4 w-4" />
        Dashboard
      </Link>

      <div className="flex items-start gap-4">
        <TemplateIcon
          name={template.icon}
          className="h-8 w-8 flex-none text-muted-foreground mt-1"
        />
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            {template.name}
          </h1>
          <p className="text-muted-foreground text-sm">
            {template.description}
          </p>
        </div>
      </div>

      {installs &&
        (added ? (
          <>
            <TemplateInstalledCard template={template} />
            <WrittenMessages template={template} />
            <SwitchToEdit template={template} />
          </>
        ) : (
          <TemplateInstallCard
            template={template}
            from="kite_bot_template"
            title="Settings"
            actionLabel={`Add ${template.name}`}
          />
        ))}

      {!!(template.sharedBotAbout ?? template.about).length && (
        <section>
          <h2 className="font-semibold mb-3">How it works</h2>
          <div className="space-y-3 text-sm text-muted-foreground">
            {(template.sharedBotAbout ?? template.about).map((paragraph, i) => (
              <p key={i}>{paragraph}</p>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

// WrittenMessages are the messages of the template the server wrote when
// adding it, like the text of a sticky message, which they can change here.
// They have no buttons with flows, which only the content would be saved of.
function WrittenMessages({ template }: { template: Template }) {
  const install = useBundleInstalls()?.find(
    (i) => i?.bundle_id === template.id
  );
  const messages = template.messages.flatMap((message) => {
    const item = install?.items.find(
      (i) => i.type === "message" && i.key === message.key && !i.deleted
    );
    return message.customize && item ? [{ ...message, id: item.id }] : [];
  });
  if (!messages.length) return null;

  return (
    <Card className="p-5 space-y-4">
      <div>
        <div className="font-semibold mb-1">Messages</div>
        <p className="text-sm text-muted-foreground">
          The bot posts the new text the next time it sends the message.
        </p>
      </div>
      {messages.map((message) => (
        <WrittenMessage
          key={message.key}
          id={message.id}
          name={message.name}
          description={message.description ?? undefined}
        />
      ))}
    </Card>
  );
}

function WrittenMessage({
  id,
  name,
  description,
}: {
  id: string;
  name: string;
  description?: string;
}) {
  const appId = useAppId();
  const message = useResponseData(useMessageQuery(appId, id));
  const mutation = useMessageDataUpdateMutation(appId);
  if (!message) return null;

  return (
    <div className="space-y-1.5">
      <div className="text-sm font-medium">{name} message</div>
      {description && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
      <MessageEditorDialog
        message={message.data}
        onClose={async (data, changed) => {
          if (!changed) return;
          const res = await mutation.mutateAsync({ messageId: id, data });
          if (!res.success) {
            toast.error(
              `Failed to save the ${name} message: ${res.error.message} (${res.error.code})`
            );
            return;
          }
          toast.success(`Saved the ${name} message`);
        }}
      >
        <Button variant="outline" className="w-full">
          Edit message
        </Button>
      </MessageEditorDialog>
    </div>
  );
}

// SwitchToEdit points to Custom setup, where the template can be changed in
// the editor, beyond its settings.
function SwitchToEdit({ template }: { template: Template }) {
  const app = useApp();
  if (!app) return null;

  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-3 rounded-xl border border-dashed p-4">
      <p className="flex-1 text-sm text-muted-foreground">
        Want to change what {template.name} does, like its replies? Switch to
        Custom setup to edit it in the editor.
      </p>
      <AppCreateDialog switchApp={app}>
        <Button variant="outline" size="sm" className="flex-none">
          Switch to Custom setup
        </Button>
      </AppCreateDialog>
    </div>
  );
}
