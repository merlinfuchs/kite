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
import { Input } from "../ui/input";
import { Plugin } from "@/lib/types/wire.gen";

export function PluginConfigureDialog({
  children,
  plugin,
}: {
  children: ReactNode;
  plugin: Plugin;
}) {
  const [dialogOpen, setDialogOpen] = useState(false);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">{plugin.name} Module</DialogTitle>
          <DialogDescription>{plugin.description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button>Import</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
