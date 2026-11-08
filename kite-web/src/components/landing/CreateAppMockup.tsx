import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

// A static copy of the Create app dialog of the dashboard.
export default function CreateAppMockup() {
  return (
    <div className="rounded-lg border bg-background p-6 space-y-4 not-prose">
      <div>
        <div className="text-lg font-semibold">Create app</div>
        <div className="text-sm text-muted-foreground">
          Kite runs your bot for you. You only need to create it in Discord
          first, which takes about a minute.
        </div>
      </div>
      <div className="space-y-2">
        <Label>Bot token</Label>
        <Input type="password" value="placeholder-token-value" readOnly />
      </div>
      <div className="flex justify-end">
        <Button tabIndex={-1}>Create app</Button>
      </div>
    </div>
  );
}
