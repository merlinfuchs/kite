import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Separator } from "../ui/separator";
import { useApp } from "@/lib/hooks/api";
import AppDeleteDialog from "./AppDeleteDialog";
import AppDuplicateDialog from "./AppDuplicateDialog";

export default function AppSettingsDangerZone() {
  const app = useApp();

  return (
    <Card className="border-destructive/50">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>
          These change your whole app, so make sure you picked the right one.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Duplicate App</div>
            <div className="text-muted-foreground">
              Copy everything you built in this app to another bot. This app
              stays as it is.
            </div>
          </div>
          {app && (
            <AppDuplicateDialog app={app}>
              <Button variant="outline" className="flex-none">
                Duplicate app
              </Button>
            </AppDuplicateDialog>
          )}
        </div>
        <Separator />
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Delete App</div>
            <div className="text-muted-foreground">
              Deleting your app will remove it from Kite and delete all
              associated data.
            </div>
          </div>
          {app && (
            <AppDeleteDialog app={app}>
              <Button
                variant="destructive"
                className="space-x-2 flex items-center flex-none"
              >
                <div>Delete app</div>
              </Button>
            </AppDeleteDialog>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
