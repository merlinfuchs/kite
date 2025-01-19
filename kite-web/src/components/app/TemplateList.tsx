import { getTemplates } from "@/lib/flow/templates";
import { useMemo } from "react";
import { TemplateListEntry } from "./TemplateListEntry";

export function TemplateList() {
  const templates = useMemo(() => getTemplates(), []);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      {templates.map((template, i) => (
        <TemplateListEntry key={i} template={template} />
      ))}
    </div>
  );
}
