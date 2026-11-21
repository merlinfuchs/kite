import { Template } from "@/lib/flow/templates";
import { useUserBundles } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { SourceAppCard } from "./BundleManageCards";
import { TemplateInstallCard } from "./TemplateInstallCard";
import { TemplateInstalledCard } from "./TemplateInstalledCard";

// AppTemplateAside is what goes next to a template's page in an app: the cards
// for adding it and for what the app has of it.
export default function AppTemplateAside({
  template,
  from,
}: {
  template: Template;
  // Where the template is added from, for analytics.
  from?: string;
}) {
  const appId = useAppId();
  const own = useUserBundles()?.find((b) => b?.id === template.id);

  return (
    <div className="space-y-5">
      {own && own.source_app_id === appId && <SourceAppCard bundle={own} />}
      <TemplateInstalledCard template={template} />
      <TemplateInstallCard template={template} from={from} />
    </div>
  );
}
