import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import TemplateIcon from "@/components/app/TemplateIcon";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Template } from "@/lib/flow/templates";
import { useApp, useAppStateGuilds, useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplates } from "@/lib/hooks/templates";
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useMemo, useState } from "react";

const breadcrumbs = [{ label: "Set up" }];

// The page apps on the Kite bot start on, after adding the bot. The query can
// pick templates and their settings, like
// ?templates=verification,giveaway&verification.role_id=123 from Xenon.
export default function AppSetupPage() {
  const appId = useAppId();
  const app = useApp();
  const guild = useAppStateGuilds()?.[0];
  const installs = useBundleInstalls();
  const templates = useTemplates();
  // The query is only there once the router is ready, after the first render.
  const { query, isReady } = useRouter();

  const picked = useMemo(
    () =>
      new Set(
        String(query.templates ?? "")
          .split(",")
          .filter(Boolean)
      ),
    [query.templates]
  );
  // Picked ones first, in the order of the Marketplace otherwise.
  const shared = useMemo(
    () =>
      (templates ?? [])
        .filter((t) => t.sharedBot)
        .sort((a, b) => Number(picked.has(b.id)) - Number(picked.has(a.id))),
    [templates, picked]
  );

  // The settings of a template in the query, by parameter key.
  function initialValues(template: Template) {
    const values: Record<string, string> = {};
    for (const p of template.parameters) {
      const value = query[`${template.id}.${p.key}`];
      if (typeof value === "string") values[p.key] = value;
    }
    return values;
  }

  return (
    <AppLayout title="Set up" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Set up the Kite bot
        </h1>
        <p className="text-muted-foreground text-sm">
          {guild
            ? `The Kite bot is in ${guild.name}. Pick the templates it should run there and fill in their settings.`
            : "Pick the templates the Kite bot should run in your server and fill in their settings."}
        </p>
      </div>
      <Separator className="my-4" />

      {!isReady || !app ? null : app.bot_type !== "shared" ? (
        <p className="text-sm text-muted-foreground">
          This app has a custom bot. Add templates to it from the{" "}
          <Link
            href={{ pathname: "/apps/[appId]/marketplace", query: { appId } }}
            className="text-primary hover:underline"
          >
            Marketplace
          </Link>
          .
        </p>
      ) : (
        <div className="space-y-3">
          {shared.map((template) => (
            <SetupTemplate
              key={template.id}
              template={template}
              added={!!installs?.some((i) => i?.bundle_id === template.id)}
              initiallyOpen={picked.has(template.id)}
              initialValues={initialValues(template)}
            />
          ))}

          <div className="pt-2">
            <Button asChild>
              <Link href={{ pathname: "/apps/[appId]", query: { appId } }}>
                Go to the dashboard
              </Link>
            </Button>
          </div>
        </div>
      )}
    </AppLayout>
  );
}

function SetupTemplate({
  template,
  added,
  initiallyOpen,
  initialValues,
}: {
  template: Template;
  added: boolean;
  initiallyOpen: boolean;
  initialValues: Record<string, string>;
}) {
  const [open, setOpen] = useState(initiallyOpen);

  return (
    <div className="space-y-3">
      <Card className="flex items-center gap-4 p-4">
        <TemplateIcon
          name={template.icon}
          className="h-6 w-6 flex-none text-muted-foreground"
        />
        <div className="min-w-0 flex-1">
          <div className="font-medium">{template.name}</div>
          <div className="text-sm text-muted-foreground">
            {template.description}
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={() => setOpen(!open)}>
          {added ? "Added" : "Set up"}
          {open ? (
            <ChevronUpIcon className="h-4 w-4 ml-1" />
          ) : (
            <ChevronDownIcon className="h-4 w-4 ml-1" />
          )}
        </Button>
      </Card>
      {open && (
        <TemplateInstallCard
          template={template}
          from="setup"
          initialValues={initialValues}
        />
      )}
    </div>
  );
}

AppSetupPage.getLayout = getAppShellLayout;
