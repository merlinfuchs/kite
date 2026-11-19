import { getKiteBotSetupLayout } from "@/components/app/KiteBotSetupShell";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import TemplateToggleRow from "@/components/app/TemplateToggleRow";
import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Template } from "@/lib/flow/templates";
import { useApp, useAppStateGuilds, useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplate, useTemplates } from "@/lib/hooks/templates";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useMemo } from "react";

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
  const router = useRouter();
  const { query } = router;

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

  // Custom bots add templates from the Marketplace.
  const custom = !!app && app.bot_type !== "shared";
  useEffect(() => {
    if (custom) {
      router.replace({
        pathname: "/apps/[appId]/marketplace",
        query: { appId },
      });
    }
  }, [custom, router, appId]);

  return (
    <BaseLayout title="Set up the Kite bot" noindex>
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

      {router.isReady && app && !custom && (
        <div className="space-y-3">
          {shared.map((template) => (
            <TemplateToggleRow
              key={template.id}
              template={template}
              label={
                installs?.some((i) => i?.bundle_id === template.id)
                  ? "Added"
                  : "Set up"
              }
              initiallyOpen={picked.has(template.id)}
            >
              {() => (
                <SetupInstall
                  id={template.id}
                  initialValues={initialValues(template)}
                />
              )}
            </TemplateToggleRow>
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
    </BaseLayout>
  );
}

function SetupInstall({
  id,
  initialValues,
}: {
  id: string;
  initialValues: Record<string, string>;
}) {
  // The list has no flows, which the parameters are found in.
  const template = useTemplate(id);
  if (!template) return null;

  return (
    <TemplateInstallCard
      template={template}
      from="setup"
      initialValues={initialValues}
    />
  );
}

AppSetupPage.getLayout = getKiteBotSetupLayout;
