import { Template } from "@/lib/flow/templates";
import { SlashSquareIcon } from "lucide-react";
import { ReactNode, useState } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardTitle } from "../ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Switch } from "../ui/switch";

export function TemplateImportDialog({
  children,
  template,
}: {
  children: ReactNode;
  template: Template;
}) {
  const [disabledCommands, setDisabledCommands] = useState<string[]>([]);

  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">{template.name}</DialogTitle>
          <DialogDescription>{template.description}</DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2">
          <div className="mb-3">
            <div className="font-medium mb-0.5">Commands</div>
            <div className="text-sm text-muted-foreground">
              Select the commands you want to import.
            </div>
          </div>
          <div className="flex flex-col gap-3">
            {template.commands.map((command, i) => (
              <Card className="px-4 pb-4 pt-3" key={i}>
                <div className="float float-right">
                  <Switch
                    checked={!disabledCommands.includes(command.name)}
                    onCheckedChange={(checked) => {
                      setDisabledCommands(
                        checked
                          ? disabledCommands.filter((c) => c !== command.name)
                          : [...disabledCommands, command.name]
                      );
                    }}
                  />
                </div>

                <div className="flex items-center gap-1.5 mb-0.5">
                  <SlashSquareIcon className="text-muted-foreground h-5 w-5" />
                  <CardTitle className="text-lg">{command.name}</CardTitle>
                </div>
                <CardDescription className="text-sm">
                  {command.description}
                </CardDescription>
              </Card>
            ))}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline">Cancel</Button>
          <Button>Import</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
