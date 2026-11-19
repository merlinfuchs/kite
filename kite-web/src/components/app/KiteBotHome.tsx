import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { hasUpdate } from "@/lib/flow/bundleInstalls";
import { Template } from "@/lib/flow/templates";
import {
  useAppFeature,
  useAppStateGuilds,
  useBundleInstalls,
  useUsageCredits,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplates } from "@/lib/hooks/templates";
import { BundleInstall } from "@/lib/types/wire.gen";
import { formatMonthlyResetDate, formatNumber } from "@/lib/utils";
import { ChevronDownIcon, ChevronUpIcon, PlusIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { Badge } from "../ui/badge";
import { Progress } from "../ui/progress";
import GetCreditsButton from "./GetCreditsButton";
import TemplateIcon from "./TemplateIcon";
import { TemplateInstalledCard } from "./TemplateInstalledCard";

// KiteBotHome is the dashboard of apps on the Kite bot: the templates it runs
// in the server, with their settings, and the credits.
export default function KiteBotHome() {
  const appId = useAppId();
  const guild = useAppStateGuilds()?.[0];
  const installs = useBundleInstalls();
  const templates = useTemplates();

  const added = (installs ?? []).flatMap((install) => {
    const template = templates?.find((t) => t.id === install?.bundle_id);
    return install && template ? [{ install, template }] : [];
  });
  const setupHref = { pathname: "/apps/[appId]/setup", query: { appId } };

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Templates</h1>
          <p className="text-muted-foreground text-sm">
            What the Kite bot does in {guild?.name ?? "your server"}. Open a
            template to change its settings.
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
        <div className="space-y-3">
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

      <CreditsCard />
    </div>
  );
}

function AddedTemplate({
  install,
  template,
}: {
  install: BundleInstall;
  template: Template;
}) {
  const [open, setOpen] = useState(false);

  return (
    <div className="space-y-3">
      <Card className="flex items-center gap-4 p-4">
        <TemplateIcon
          name={template.icon}
          className="h-6 w-6 flex-none text-muted-foreground"
        />
        <div className="min-w-0 flex-1">
          <div className="font-medium flex items-center gap-2">
            {template.name}
            {hasUpdate(install) && <Badge variant="secondary">Update</Badge>}
          </div>
          <div className="text-sm text-muted-foreground">
            {template.description}
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={() => setOpen(!open)}>
          Settings
          {open ? (
            <ChevronUpIcon className="h-4 w-4 ml-1" />
          ) : (
            <ChevronDownIcon className="h-4 w-4 ml-1" />
          )}
        </Button>
      </Card>
      {open && <TemplateInstalledCard template={template} />}
    </div>
  );
}

function CreditsCard() {
  const credits = useUsageCredits();
  const creditsPerMonth = useAppFeature((f) => f.usage_credits_per_month);
  if (!credits || !creditsPerMonth) return null;

  return (
    <Card className="p-5">
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-4">
        <div>
          <div className="font-medium">
            {formatNumber(credits.credits_used)} of{" "}
            {formatNumber(creditsPerMonth)} credits used
          </div>
          <p className="text-sm text-muted-foreground">
            Sending a message or giving a role costs 1 credit. Credits reset on{" "}
            {formatMonthlyResetDate()}.
          </p>
        </div>
        <GetCreditsButton from="kite_bot_home" />
      </div>
      <Progress
        value={Math.min(100, (credits.credits_used / creditsPerMonth) * 100)}
      />
    </Card>
  );
}
