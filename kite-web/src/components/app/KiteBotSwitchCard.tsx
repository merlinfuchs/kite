import { useApp } from "@/lib/hooks/api";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import AppCreateDialog from "./AppCreateDialog";

// KiteBotSwitchCard offers an app on the Kite bot to move to a bot of its
// own, Custom setup.
export default function KiteBotSwitchCard() {
  const app = useApp();
  if (!app) return null;

  return (
    <Card id="switch" className="p-6 scroll-mt-20">
      <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="font-bold pb-1">Switch to Custom setup</div>
          <div className="text-muted-foreground text-sm">
            Use a bot of your own instead of the Kite bot. You get the editor to
            change your templates and build commands, your own status, and your
            bot&apos;s name and avatar everywhere. The app keeps its templates,
            stored variables, logs, and plan. You become its owner, and everyone
            else, also its current owner, loses access until you add them as
            collaborators. Whoever pays for its plan keeps access to manage it.
            The Kite bot stays in your server until you remove it. Messages it
            posted, like the verification message, then need to be posted again
            by your bot, from the template&apos;s page.
          </div>
        </div>
        <AppCreateDialog switchApp={app}>
          <Button variant="secondary" className="flex-none">
            Switch
          </Button>
        </AppCreateDialog>
      </div>
    </Card>
  );
}
