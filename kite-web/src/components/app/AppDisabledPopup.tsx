import { useApp } from "@/lib/hooks/api";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { OctagonAlertIcon } from "lucide-react";

export default function AppDisabledPopup() {
  const app = useApp();

  if (app?.enabled || !app?.disabled_reason) return null;

  return (
    <Card className="shadow-md max-w-96 fixed top-5 right-5 ml-5 z-50 border-red-500">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-base flex items-center gap-2">
          <OctagonAlertIcon className="w-5 h-5 text-red-500" />
          App Disabled
        </CardTitle>
        <CardDescription className="text-sm">
          Your app is currently disabled: {app.disabled_reason}
        </CardDescription>
      </CardHeader>
    </Card>
  );
}
