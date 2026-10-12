import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  decodePermissionsBitset,
  PermissionCategory,
  permissionBits,
} from "@/lib/discord/permissions";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getApiUrl } from "@/lib/api/client";
import { useAppsQuery, useUserQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { App } from "@/lib/types/wire.gen";
import CopyButton from "@/tools/common/components/CopyButton";
import { ExternalLinkIcon } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

// In this order the two columns end up about the same height.
const categories = (
  ["General", "Voice", "Text", "Members", "Advanced"] as PermissionCategory[]
).map((name) => ({
  name,
  permissions: permissionBits.filter((p) => p.category === name),
}));

// Comes back to the calculator after signing in.
const loginUrl = getApiUrl(
  "/v1/auth/login?redirect=/tools/permissions-calculator"
);

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

  // Signed in users can pick one of their Kite apps instead of pasting its ID.
  const userQuery = useUserQuery();
  const user = useResponseData(userQuery);
  const apps = useResponseData(useAppsQuery(!!user))?.filter(
    (app): app is App => !!app
  );

  // Shows the integer at the bottom of the screen while the permissions are
  // visible but its input has scrolled away.
  const inputRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const [inputVisible, setInputVisible] = useState(true);
  const [listVisible, setListVisible] = useState(false);
  useEffect(() => {
    const observer = new IntersectionObserver((entries) => {
      for (const entry of entries) {
        if (entry.target === inputRef.current)
          setInputVisible(entry.isIntersecting);
        if (entry.target === listRef.current)
          setListVisible(entry.isIntersecting);
      }
    });
    observer.observe(inputRef.current!);
    observer.observe(listRef.current!);
    return () => observer.disconnect();
  }, []);

  const enabled = useMemo(
    () => new Set<number>(decodePermissionsBitset(value).map((p) => p.bit)),
    [value]
  );

  // Changes only this bit, so bits of permissions we don't know yet survive.
  function toggle(bit: number, checked: boolean) {
    const bits = BigInt(value || "0");
    const mask = BigInt(1) << BigInt(bit);
    setValue((checked ? bits | mask : bits & ~mask).toString());
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
      <div ref={inputRef} className="space-y-2">
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

      <div ref={listRef} className="md:columns-2 gap-4">
        {categories.map((category) => (
          <div
            key={category.name}
            className="rounded-lg bg-muted/50 p-4 mb-4 break-inside-avoid"
          >
            <div className="flex items-center justify-between mb-3">
              <h2 className="font-semibold">{category.name}</h2>
              <span className="text-xs text-muted-foreground">
                {category.permissions.filter((p) => enabled.has(p.bit)).length}{" "}
                of {category.permissions.length}
              </span>
            </div>
            <div className="space-y-2">
              {category.permissions.map((p) => (
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

      <div className="border-t pt-8 space-y-5">
        <div>
          <h2 className="text-xl font-semibold">Invite Link</h2>
          <p className="text-sm text-muted-foreground">
            Enter the application ID of your bot to get an invite link with
            these permissions. You find it on the General Information page of
            your app in the Discord Developer Portal.
          </p>
        </div>
        {!!apps?.length && (
          <div className="space-y-2">
            <Label>Your Kite bots</Label>
            <Select
              value={
                apps.some((app) => app.discord_id === clientId) ? clientId : ""
              }
              onValueChange={setClientId}
            >
              <SelectTrigger>
                <SelectValue placeholder="Pick a bot" />
              </SelectTrigger>
              <SelectContent>
                {apps.map((app) => (
                  <SelectItem key={app.id} value={app.discord_id}>
                    {app.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
        {!userQuery.isPending && !user && (
          <p className="text-sm text-muted-foreground">
            Built your bot with Kite?{" "}
            <a href={loginUrl} className="text-primary hover:underline">
              Sign in
            </a>{" "}
            to pick it from a list.
          </p>
        )}
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

      {!inputVisible && listVisible && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 rounded-full border bg-background shadow-lg pl-5 pr-2 py-2">
          <span className="text-sm text-muted-foreground">Permissions</span>
          <span className="font-mono">{value || "0"}</span>
          <CopyButton value={value || "0"} />
        </div>
      )}
    </div>
  );
}
