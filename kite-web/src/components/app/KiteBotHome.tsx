import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { hasUpdate } from "@/lib/flow/bundleInstalls";
import { Template } from "@/lib/flow/templates";
import {
  useApp,
  useAppStateGuildChannels,
  useAppStateGuildRoles,
  useAppStateGuilds,
  useBundleInstalls,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useFullTemplates } from "@/lib/hooks/templates";
import { BundleInstall } from "@/lib/types/wire.gen";
import { ChevronRightIcon, PlusIcon } from "lucide-react";
import Link from "next/link";
import { Badge } from "../ui/badge";
import KiteBotLook from "./KiteBotLook";
import KiteBotStatus from "./KiteBotStatus";
import TemplateIcon from "./TemplateIcon";

// KiteBotHome is the dashboard of apps on the Kite bot: whether the bot runs
// fine, and the templates it runs in the server, with their settings.
export default function KiteBotHome() {
  const appId = useAppId();
  const guild = useAppStateGuilds()?.[0];
  const installs = useBundleInstalls();
  // With their flows, which the permissions and updates are worked out from.
  const templates = useFullTemplates(
    (installs ?? []).flatMap((i) => (i ? [i.bundle_id] : []))
  );

  const added = (installs ?? []).flatMap((install) => {
    const template = templates?.find((t) => t.id === install?.bundle_id);
    return install && template ? [{ install, template }] : [];
  });
  const setupHref = { pathname: "/apps/[appId]/setup", query: { appId } };

  return (
    <div className="space-y-8 max-w-3xl">
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          {guild?.name ?? "Your server"}
        </h1>
        <p className="text-muted-foreground text-sm">
          What the Kite bot does in your server, and whether it works.
        </p>
      </div>

      <KiteBotLook />

      {installs && templates && <KiteBotStatus added={added} />}

      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold mb-1">Templates</h2>
          <p className="text-muted-foreground text-sm">
            Open a template to see how it works and change its settings.
          </p>
        </div>
        {!!added.length && (
          <Button asChild className="flex-none">
            <Link href={setupHref}>
              <PlusIcon className="h-4 w-4 mr-1" />
              Add templates
            </Link>
          </Button>
        )}
      </div>

      {installs && templates && (
        <div className="space-y-3 !mt-4">
          {added.length ? (
            added.map(({ install, template }) => (
              <AddedTemplate
                key={install.id}
                install={install}
                template={template}
              />
            ))
          ) : (
            <Card className="p-8 text-center">
              <div className="font-medium mb-1">No templates yet</div>
              <p className="text-sm text-muted-foreground mb-4">
                Add one and the Kite bot starts running it in your server.
              </p>
              <Button asChild>
                <Link href={setupHref}>Add templates</Link>
              </Button>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}

// AddedTemplate links to the page of an added template, with its settings in
// short, like the channel it posts in.
function AddedTemplate({
  install,
  template,
}: {
  install: BundleInstall;
  template: Template;
}) {
  const appId = useAppId();
  const guildId = useApp()?.discord_guild_id ?? null;
  const channels = useAppStateGuildChannels(guildId);
  const roles = useAppStateGuildRoles(guildId);

  const settings = template.parameters.flatMap((p) => {
    const value = install.parameters?.[p.key];
    if (!value) return [];
    if (p.type === "channel" || p.type === "category") {
      const channel = channels?.find((c) => c?.id === value);
      return channel ? [`#${channel.name}`] : [];
    }
    if (p.type === "role") {
      const role = roles?.find((r) => r?.id === value);
      return role ? [`@${role.name}`] : [];
    }
    return [];
  });

  return (
    <Link
      href={{
        pathname: "/apps/[appId]/bundles/[bundleId]",
        query: { appId, bundleId: template.id },
      }}
      className="block rounded-xl outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <Card className="flex items-center gap-4 p-4 transition-colors hover:border-foreground/30">
        <TemplateIcon
          name={template.icon}
          className="h-6 w-6 flex-none text-muted-foreground"
        />
        <div className="min-w-0 flex-1">
          <div className="font-medium flex items-center gap-2">
            {template.name}
            {hasUpdate(install) && <Badge variant="secondary">Update</Badge>}
          </div>
          <div className="text-sm text-muted-foreground truncate">
            {settings.length ? settings.join(", ") : template.description}
          </div>
        </div>
        <ChevronRightIcon className="h-5 w-5 flex-none text-muted-foreground" />
      </Card>
    </Link>
  );
}
