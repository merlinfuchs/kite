import { prepareTemplateFlow, Template } from "@/lib/flow/templates";
import { SatelliteDishIcon, SlashSquareIcon } from "lucide-react";
import { ReactNode, useCallback, useState } from "react";
import { Button } from "../ui/button";
import { Card, CardDescription, CardTitle } from "../ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Switch } from "../ui/switch";
import {
  useCommandsImportMutation,
  useEventListenersImportMutation,
} from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { Label } from "../ui/label";
import { Input } from "../ui/input";

export function TemplateImportDialog({
  children,
  template,
}: {
  children: ReactNode;
  template: Template;
}) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const [disabledCommands, setDisabledCommands] = useState<number[]>([]);
  const [disabledEventListeners, setDisabledEventListeners] = useState<
    number[]
  >([]);

  const appId = useAppId();

  const commandsImportMutation = useCommandsImportMutation(appId);
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

  const [inputValues, setInputValues] = useState<Record<string, string>>({});

  const handleImport = useCallback(() => {
    for (const input of template.inputs) {
      if (input.required && !inputValues[input.key]) {
        toast.error(`${input.label} is required`);
        return;
      }
    }

    const commands = template.commands
      .filter((_, i) => !disabledCommands.includes(i))
      .map((c) => ({
        enabled: true,
        flow_source: prepareTemplateFlow(c.flowSource(inputValues)),
      }));
    const eventListeners = template.eventListeners
      .filter((_, i) => !disabledEventListeners.includes(i))
      .map((c) => ({
        enabled: true,
        source: c.source,
        flow_source: prepareTemplateFlow(c.flowSource(inputValues)),
      }));

    if (commands.length != 0) {
      commandsImportMutation.mutate(
        { commands },
        {
          onSuccess: (res) => {
            if (res.success) {
              toast.success(
                `${commands.length} commands imported successfully`
              );
            } else {
              toast.error(
                `Failed to import commands: ${res.error.message} (${res.error.code})`
              );
            }
          },
        }
      );
    }
    if (eventListeners.length != 0) {
      eventListenersImportMutation.mutate(
        { event_listeners: eventListeners },
        {
          onSuccess: (res) => {
            if (res.success) {
              toast.success(
                `${eventListeners.length} event listeners imported successfully`
              );
            } else {
              toast.error(
                `Failed to import event listeners: ${res.error.message} (${res.error.code})`
              );
            }
          },
        }
      );
    }

    setDialogOpen(false);
  }, [
    template,
    disabledCommands,
    disabledEventListeners,
    inputValues,
    commandsImportMutation,
    eventListenersImportMutation,
  ]);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">{template.name} Template</DialogTitle>
          <DialogDescription>{template.description}</DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2 space-y-5">
          {!!template.inputs.length && (
            <div className="flex flex-col gap-3">
              {template.inputs.map((input) => (
                <div key={input.key}>
                  <div className="mb-3">
                    <div className="flex items-center gap-2">
                      <div className="font-medium mb-0.5">{input.label}</div>
                      <div className="text-xs text-muted-foreground">
                        {input.required ? (
                          <span className="text-red-500">required</span>
                        ) : (
                          <span className="text-muted-foreground">
                            optional
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {input.description}
                    </div>
                  </div>
                  <Input
                    value={inputValues[input.key] || ""}
                    onChange={(e) =>
                      setInputValues({
                        ...inputValues,
                        [input.key]: e.target.value,
                      })
                    }
                  />
                </div>
              ))}
            </div>
          )}

          {!!template.commands.length && (
            <div>
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
                        checked={!disabledCommands.includes(i)}
                        onCheckedChange={(checked) => {
                          setDisabledCommands(
                            checked
                              ? disabledCommands.filter((c) => c !== i)
                              : [...disabledCommands, i]
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
          )}

          {!!template.eventListeners.length && (
            <div>
              <div className="mb-3">
                <div className="font-medium mb-0.5">Event Listeners</div>
                <div className="text-sm text-muted-foreground">
                  Select the event listeners you want to import.
                </div>
              </div>
              <div className="flex flex-col gap-3">
                {template.eventListeners.map((listener, i) => (
                  <Card className="px-4 pb-4 pt-3" key={i}>
                    <div className="float float-right">
                      <Switch
                        checked={!disabledEventListeners.includes(i)}
                        onCheckedChange={(checked) => {
                          setDisabledEventListeners(
                            checked
                              ? disabledEventListeners.filter((c) => c !== i)
                              : [...disabledEventListeners, i]
                          );
                        }}
                      />
                    </div>

                    <div className="flex items-center gap-1.5 mb-0.5">
                      <SatelliteDishIcon className="text-muted-foreground h-5 w-5" />
                      <CardTitle className="text-lg">{listener.type}</CardTitle>
                    </div>
                    <CardDescription className="text-sm">
                      {listener.description}
                    </CardDescription>
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={handleImport}>Import</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
