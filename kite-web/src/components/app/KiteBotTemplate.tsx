import { Template } from "@/lib/flow/templates";
import { useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { ArrowLeftIcon } from "lucide-react";
import Link from "next/link";
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
          <TemplateInstalledCard template={template} />
        ) : (
          <TemplateInstallCard
            template={template}
            from="kite_bot_template"
            title="Settings"
            actionLabel={`Add ${template.name}`}
          />
        ))}

      {!!template.about.length && (
        <section>
          <h2 className="font-semibold mb-3">How it works</h2>
          <div className="space-y-3 text-sm text-muted-foreground">
            {template.about.map((paragraph, i) => (
              <p key={i}>{paragraph}</p>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
