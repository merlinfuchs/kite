import { useApp } from "@/lib/hooks/api";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { OctagonAlertIcon } from "lucide-react";
import { Button } from "../ui/button";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";

export default function AppDisabledPopup() {
  const app = useApp();

  if (app?.enabled || !app?.disabled_reason) return null;

  return (
    <Card className="shadow-md max-w-96 fixed inset-x-3 ml-auto top-16 md:inset-x-auto md:top-5 md:right-5 z-50 border-red-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2">
          <OctagonAlertIcon className="w-5 h-5 text-red-500" />
          {app.token_invalid
            ? "Your Bot Token Stopped Working"
            : "App Disabled"}
        </CardTitle>
        <CardDescription className="text-sm">
          {app.token_invalid
            ? "The token was probably reset in the Discord Developer Portal. Update it to start your bot again."
            : `Your app is currently disabled: ${app.disabled_reason}`}
        </CardDescription>
        {app.token_invalid && (
          <div className="pt-2">
            <AppTokenUpdateDialog app={app} from="popup">
              <Button size="sm">Update token</Button>
            </AppTokenUpdateDialog>
          </div>
        )}
      </CardHeader>
    </Card>
  );
}
