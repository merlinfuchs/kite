import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useApp } from "@/lib/hooks/api";
import AppTokenUpdateDialog from "./AppTokenUpdateDialog";

export default function AppSettingsCredentials() {
  const app = useApp();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Credentials</CardTitle>
        <CardDescription>
          Update your bot&apos;s token if you reset it in the Discord Developer
          Portal.
        </CardDescription>
      </CardHeader>
      <CardFooter className="border-t px-6 py-4">
        {app && (
          <AppTokenUpdateDialog app={app} from="settings">
            <Button>Update token</Button>
          </AppTokenUpdateDialog>
        )}
      </CardFooter>
    </Card>
  );
}
