import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { useKiteSettings } from "@/lib/hooks/useKiteSettings";
import { MobileNavStyle, MobileNavTabs } from "@/lib/settings/types";
import { Undo2Icon } from "lucide-react";

export default function MobileNavigationSection() {
  const { settings, updateSettings, resetSettings } = useKiteSettings();

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="mobile-nav-style" className="text-sm font-medium">
            Bar Layout
          </Label>
          <p className="text-xs text-muted-foreground">
            Choose between an edge-to-edge docked bar or a floating pill dock.
          </p>
        </div>
        <Select
          value={settings.mobileNavStyle}
          onValueChange={(val: MobileNavStyle) =>
            updateSettings({ mobileNavStyle: val })
          }
        >
          <SelectTrigger id="mobile-nav-style" className="w-[180px]">
            <SelectValue placeholder="Select style" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="docked">Standard Docked</SelectItem>
            <SelectItem value="floating">Floating Island</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="mobile-nav-tabs" className="text-sm font-medium">
            Visible Tabs
          </Label>
          <p className="text-xs text-muted-foreground">
            Choose which action shortcuts appear directly on the bottom bar.
          </p>
        </div>
        <Select
          value={settings.mobileNavTabs}
          onValueChange={(val: MobileNavTabs) =>
            updateSettings({ mobileNavTabs: val })
          }
        >
          <SelectTrigger id="mobile-nav-tabs" className="w-[195px]">
            <SelectValue placeholder="Select tabs" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="3-tabs">3 Tabs (Dash, Cmds, Menu)</SelectItem>
            <SelectItem value="4-tabs">4 Tabs (+ Listeners)</SelectItem>
            <SelectItem value="5-tabs">5 Tabs (+ Listeners, Logs)</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="show-labels" className="text-sm font-medium">
            Show Text Labels
          </Label>
          <p className="text-xs text-muted-foreground">
            Display titles below icons (e.g. Dashboard, Commands, Listeners).
          </p>
        </div>
        <Switch
          id="show-labels"
          checked={settings.showMobileLabels}
          onCheckedChange={(checked) =>
            updateSettings({ showMobileLabels: checked })
          }
        />
      </div>

      <div className="flex justify-end pt-2">
        <Button
          variant="outline"
          size="sm"
          onClick={resetSettings}
          className="gap-1.5 text-xs"
        >
          <Undo2Icon className="size-3.5" />
          Reset Mobile Settings
        </Button>
      </div>
    </div>
  );
}
