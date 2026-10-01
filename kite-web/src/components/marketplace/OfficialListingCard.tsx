import { BadgeCheckIcon, TriangleAlertIcon } from "lucide-react";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { TemplateImportDialog } from "../app/TemplateImportDialog";
import { Template } from "@/lib/flow/templates";
import {
  getRiskyBlocks,
  listingKind,
  listingKindLabel,
  listingSummary,
} from "@/lib/marketplace";
import { useMemo } from "react";

// Official templates shown like community listings, so the tabs look the same.
export default function OfficialListingCard({
  template,
}: {
  template: Template;
}) {
  const counts = {
    command_count: template.commands.length,
    event_listener_count: template.eventListeners.length,
    message_count: 0,
  };
  const kind = listingKind(counts);

  const risky = useMemo(() => {
    const blockTypes = new Set<string>();
    for (const item of [...template.commands, ...template.eventListeners]) {
      try {
        for (const node of item.flowSource({}).nodes) {
          if (node.type) blockTypes.add(node.type);
        }
      } catch {
        // Some flows need their inputs, the import dialog shows them anyway.
      }
    }
    return getRiskyBlocks(Array.from(blockTypes)).length > 0;
  }, [template]);

  return (
    <TemplateImportDialog template={template}>
      <Card
        className="cursor-pointer hover:border-primary/60 transition-colors flex flex-col"
        role="button"
      >
        <CardHeader className="flex flex-row gap-4 p-4 space-y-0 flex-1">
          <div className="h-10 w-10 bg-primary/40 flex-none rounded-md flex items-center justify-center">
            <template.icon className="w-6 h-6 text-primary" />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 mb-1">
              <CardTitle className="text-base truncate">
                {template.name}
              </CardTitle>
            </div>
            <CardDescription className="line-clamp-2 mb-3">
              {template.description}
            </CardDescription>
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="outline">{listingKindLabel(kind)}</Badge>
              {kind === "module" && (
                <span className="text-xs text-muted-foreground">
                  {listingSummary(counts)}
                </span>
              )}
              {risky && (
                <TriangleAlertIcon
                  className="h-4 w-4 text-yellow-500"
                  aria-label="Uses sensitive blocks"
                />
              )}
            </div>
          </div>
        </CardHeader>
        <div className="flex items-center justify-between px-4 pb-4 gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <BadgeCheckIcon className="h-5 w-5 text-primary flex-none" />
            <span className="text-sm text-muted-foreground truncate">
              Kite Team
            </span>
          </div>
          <Badge variant="secondary" className="flex-none">
            Official
          </Badge>
        </div>
      </Card>
    </TemplateImportDialog>
  );
}
