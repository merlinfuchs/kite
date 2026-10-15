import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

// A static copy of the Create App dialog of the dashboard.
export default function CreateAppMockup() {
  return (
    <div className="rounded-lg border bg-background p-6 space-y-4 not-prose">
      <div>
        <div className="text-lg font-semibold">Create App</div>
        <div className="text-sm text-muted-foreground">
          Create a new app to get started with Kite and start building your app!
        </div>
      </div>
      <div className="space-y-2">
        <Label>Bot Token</Label>
        <Input type="password" value="placeholder-token-value" readOnly />
      </div>
      <div className="flex justify-end">
        <Button tabIndex={-1}>Create app</Button>
      </div>
    </div>
  );
}
