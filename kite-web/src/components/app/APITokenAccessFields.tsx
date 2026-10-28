import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Control,
  FieldValues,
  Path,
  useFormContext,
  useWatch,
} from "react-hook-form";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Skeleton } from "../ui/skeleton";
import { useAppsQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { cn } from "@/lib/utils";

export interface Option {
  value: string;
  label: string;
  // How tokens with it are described in the list.
  summary?: string;
}

export const scopeOptions: Option[] = [
  {
    value: "kite:write",
    label: "Full access",
    summary: "Read and change apps",
  },
  { value: "kite:read", label: "Read only", summary: "Only read apps" },
];

// Write first, like scopeOptions, so tokens are described by their highest.
export const discordScopeOptions: Option[] = [
  {
    value: "discord:write",
    label: "Full access",
    summary: "Read and change Discord",
  },
  { value: "discord:read", label: "Read only", summary: "Read Discord" },
  { value: "none", label: "None" },
];

const appsOptions: Option[] = [
  { value: "all", label: "All apps" },
  { value: "some", label: "Some apps" },
];

// What a token or OAuth client can do, picked by the user.
export interface AccessFields {
  scope: string;
  // "none" for no access to the Discord API.
  discordScope: string;
  apps: string;
  appIds: string[];
}

// The scopes of the access picked in the fields.
export function accessScopes(data: AccessFields) {
  return [data.scope, data.discordScope].filter((s) => s !== "none");
}

// APITokenAccessFields picks the access of a token, in a form with the fields
// of AccessFields. With scopes, it only offers those, like the ones an OAuth
// client asked for.
export default function APITokenAccessFields({
  scopes,
}: {
  scopes?: string[];
}) {
  const { control } = useFormContext<AccessFields>();
  const allowed = (o: Option) =>
    !scopes || o.value === "none" || scopes.includes(o.value);
  const someApps = useWatch({ control, name: "apps" }) === "some";
  const discordOptions = discordScopeOptions.filter(allowed);
  // Without any Discord scope to offer, like for OAuth clients, there's
  // nothing to pick.
  const showDiscord = discordOptions.length > 1;

  return (
    <>
      <div
        className={cn(
          "grid gap-2",
          showDiscord ? "sm:grid-cols-3" : "sm:grid-cols-2"
        )}
      >
        <SelectField
          control={control}
          name="scope"
          label="Apps access"
          showLabel
          options={scopeOptions.filter(allowed)}
        />
        {showDiscord && (
          <SelectField
            control={control}
            name="discordScope"
            label="Discord API"
            showLabel
            options={discordOptions}
          />
        )}
        <SelectField
          control={control}
          name="apps"
          label="Which apps"
          showLabel
          options={appsOptions}
        />
      </div>
      {someApps && <AppPicker control={control} />}
    </>
  );
}

function AppPicker({ control }: { control: Control<AccessFields> }) {
  const apps = useResponseData(useAppsQuery());

  return (
    <FormField
      control={control}
      name="appIds"
      rules={{
        validate: (v) => v.length > 0 || "Pick at least one app",
      }}
      render={({ field }) => (
        <FormItem className="basis-full">
          <div className="grid max-h-40 gap-2 overflow-y-auto rounded-md border p-3">
            {!apps ? (
              <Skeleton className="h-5" />
            ) : apps.length === 0 ? (
              <div className="text-sm text-muted-foreground">No apps yet.</div>
            ) : (
              apps.map(
                (app) =>
                  app && (
                    <label
                      key={app.id}
                      className="flex items-center gap-2 text-sm"
                    >
                      <Checkbox
                        checked={field.value.includes(app.id)}
                        onCheckedChange={(checked) =>
                          field.onChange(
                            checked
                              ? [...field.value, app.id]
                              : field.value.filter((id) => id !== app.id)
                          )
                        }
                      />
                      <span className="truncate">{app.name}</span>
                    </label>
                  )
              )
            )}
          </div>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

export function SelectField<T extends FieldValues>({
  control,
  name,
  label,
  options,
  showLabel,
  className,
  triggerClassName,
}: {
  control: Control<T>;
  name: Path<T>;
  label: string;
  // Above the select, otherwise it's only for screen readers.
  showLabel?: boolean;
  options: Option[];
  className?: string;
  triggerClassName?: string;
}) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          {showLabel && (
            <FormLabel className="text-xs text-muted-foreground">
              {label}
            </FormLabel>
          )}
          <Select value={field.value} onValueChange={field.onChange}>
            <FormControl>
              <SelectTrigger
                className={cn("w-full", triggerClassName)}
                aria-label={label}
              >
                <SelectValue />
              </SelectTrigger>
            </FormControl>
            <SelectContent>
              {options.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
