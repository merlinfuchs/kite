import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  decodePermissionsBitset,
  encodePermissionsBitset,
  PermissionCategory,
  permissionBits,
} from "@/lib/discord/permissions";
import CopyButton from "@/tools/common/components/CopyButton";
import { ExternalLinkIcon } from "lucide-react";
import { useMemo, useState } from "react";

const categories: PermissionCategory[] = [
  "General",
  "Members",
  "Text",
  "Voice",
  "Advanced",
];

const scopes = [
  { id: "bot", label: "bot", description: "Adds the bot to the server" },
  {
    id: "applications.commands",
    label: "applications.commands",
    description: "Lets the bot create slash commands",
  },
];

export default function PermissionsCalculator() {
  const [value, setValue] = useState("0");
  const [clientId, setClientId] = useState("");
  const [selectedScopes, setSelectedScopes] = useState([
    "bot",
    "applications.commands",
  ]);

  const enabled = useMemo(
    () => new Set<number>(decodePermissionsBitset(value).map((p) => p.bit)),
    [value]
  );

  function toggle(bit: number, checked: boolean) {
    const bits = new Set(enabled);
    if (checked) bits.add(bit);
    else bits.delete(bit);
    setValue(encodePermissionsBitset(Array.from(bits)));
  }

  const validClientId = /^\d{17,20}$/.test(clientId);
  const inviteUrl = useMemo(() => {
    const params = new URLSearchParams({ client_id: clientId });
    if (selectedScopes.includes("bot")) params.set("permissions", value || "0");
    params.set("integration_type", "0");
    params.set("scope", selectedScopes.join(" "));
    return `https://discord.com/oauth2/authorize?${params}`;
  }, [clientId, selectedScopes, value]);

  return (
    <div className="space-y-10">
      <div className="rounded-lg border p-5 space-y-2">
        <Label htmlFor="permissions">Permissions integer</Label>
        <div className="flex gap-3">
          <Input
            id="permissions"
            value={value}
            onChange={(e) => setValue(e.target.value.replace(/\D/g, ""))}
            className="font-mono"
          />
          <CopyButton value={value || "0"} />
          <Button
            variant="outline"
            size="sm"
            className="flex-none"
            onClick={() => setValue("0")}
          >
            Clear
          </Button>
        </div>
        <p className="text-sm text-muted-foreground">
          Select permissions below, or paste an integer to see which permissions
          it contains.
        </p>
      </div>

      {enabled.has(3) && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm">
          Administrator gives the bot every permission and lets it ignore
          channel overwrites. Only ask for it if your bot really needs it.
        </div>
      )}

      <div className="grid md:grid-cols-2 gap-x-10 gap-y-8">
        {categories.map((category) => (
          <div key={category}>
            <h2 className="font-semibold mb-3">{category}</h2>
            <div className="space-y-2">
              {permissionBits
                .filter((p) => p.category === category)
                .map((p) => (
                  <label
                    key={p.bit}
                    className="flex items-center gap-3 text-sm cursor-pointer"
                  >
                    <Checkbox
                      checked={enabled.has(p.bit)}
                      onCheckedChange={(checked) => toggle(p.bit, !!checked)}
                    />
                    {p.label}
                  </label>
                ))}
            </div>
          </div>
        ))}
      </div>

      <div className="rounded-lg border p-5 space-y-5">
        <div>
          <h2 className="text-xl font-semibold">Invite Link</h2>
          <p className="text-sm text-muted-foreground">
            Enter the application ID of your bot to get an invite link with
            these permissions. You find it on the General Information page of
            your app in the Discord Developer Portal.
          </p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="client-id">Application ID</Label>
          <Input
            id="client-id"
            placeholder="123456789012345678"
            value={clientId}
            onChange={(e) => setClientId(e.target.value.trim())}
            className="font-mono"
          />
          {clientId && !validClientId && (
            <p className="text-sm text-destructive">
              An application ID is a number with 17 to 20 digits.
            </p>
          )}
        </div>
        <div className="space-y-2">
          <Label>Scopes</Label>
          {scopes.map((scope) => (
            <label
              key={scope.id}
              className="flex items-center gap-3 text-sm cursor-pointer"
            >
              <Checkbox
                checked={selectedScopes.includes(scope.id)}
                onCheckedChange={(checked) =>
                  setSelectedScopes(
                    checked
                      ? [...selectedScopes, scope.id]
                      : selectedScopes.filter((s) => s !== scope.id)
                  )
                }
              />
              <span className="font-mono">{scope.label}</span>
              <span className="text-muted-foreground">{scope.description}</span>
            </label>
          ))}
        </div>
        {validClientId && selectedScopes.length > 0 && (
          <div className="space-y-2">
            <Label>Your invite link</Label>
            <div className="flex gap-3">
              <Input value={inviteUrl} readOnly className="font-mono" />
              <CopyButton value={inviteUrl} />
              <Button asChild size="sm" className="flex-none">
                <a href={inviteUrl} target="_blank" rel="noreferrer noopener">
                  Open
                  <ExternalLinkIcon className="w-4 h-4 ml-2" />
                </a>
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
