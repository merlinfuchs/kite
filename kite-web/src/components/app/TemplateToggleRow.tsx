import { Template } from "@/lib/flow/templates";
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react";
import { ReactNode, useState } from "react";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import TemplateIcon from "./TemplateIcon";

// TemplateToggleRow is a template in a list that opens to show children, like
// its settings.
export default function TemplateToggleRow({
  template,
  label,
  badge,
  initiallyOpen = false,
  children,
}: {
  template: Pick<Template, "icon" | "name" | "description">;
  label: string;
  badge?: ReactNode;
  initiallyOpen?: boolean;
  // Only rendered while it's open.
  children: () => ReactNode;
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
          <div className="font-medium flex items-center gap-2">
            {template.name}
            {badge}
          </div>
          <div className="text-sm text-muted-foreground">
            {template.description}
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={() => setOpen(!open)}>
          {label}
          {open ? (
            <ChevronUpIcon className="h-4 w-4 ml-1" />
          ) : (
            <ChevronDownIcon className="h-4 w-4 ml-1" />
          )}
        </Button>
      </Card>
      {open && children()}
    </div>
  );
}
