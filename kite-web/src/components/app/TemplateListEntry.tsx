import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Template } from "@/lib/flow/templates";
import { TemplateImportDialog } from "./TemplateImportDialog";

export function TemplateListEntry({ template }: { template: Template }) {
  return (
    <Card>
      <CardHeader className="flex flex-row gap-4 p-4">
        <div className="h-10 w-10 bg-primary/40 flex-none rounded-md flex items-center justify-center">
          <template.icon className="w-6 h-6 text-primary" />
        </div>
        <div>
          <CardTitle className="mb-2">{template.name}</CardTitle>
          <CardDescription>{template.description}</CardDescription>
        </div>
      </CardHeader>
      <CardFooter className="p-4 pt-1 flex justify-end">
        <TemplateImportDialog template={template}>
          <Button variant="outline">View details</Button>
        </TemplateImportDialog>
      </CardFooter>
    </Card>
  );
}
