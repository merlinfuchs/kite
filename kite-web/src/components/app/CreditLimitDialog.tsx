import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ReactNode, useState } from "react";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { useForm } from "react-hook-form";
import {
  useCreditLimitCreateMutation,
  useCreditLimitUpdateMutation,
} from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { CreditLimit } from "@/lib/types/wire.gen";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Switch } from "../ui/switch";
import { useAppStateGuilds } from "@/lib/hooks/api";

interface FormFields {
  scope: string;
  target: "default" | "specific";
  target_id: string;
  period: string;
  unlimited: boolean;
  credits: string;
}

export interface CreditLimitDefaults {
  scope?: string;
  target_id?: string;
  period?: string;
}

function defaultValues(
  limit?: CreditLimit,
  defaults?: CreditLimitDefaults
): FormFields {
  if (limit) {
    return {
      scope: limit.scope,
      target: limit.target_id ? "specific" : "default",
      target_id: limit.target_id ?? "",
      period: limit.period,
      unlimited: limit.credits === null,
      credits: limit.credits?.toString() ?? "",
    };
  }

  return {
    scope: defaults?.scope ?? "guild",
    target: defaults?.target_id ? "specific" : "default",
    target_id: defaults?.target_id ?? "",
    period: defaults?.period ?? "day",
    unlimited: false,
    credits: "",
  };
}

// Creates a credit limit, or edits one when given. Defaults prefill the form,
// e.g. when setting a limit for a server from the usage table.
export default function CreditLimitDialog({
  children,
  limit,
  defaults,
}: {
  children: ReactNode;
  limit?: CreditLimit;
  defaults?: CreditLimitDefaults;
}) {
  const [open, setOpen] = useState(false);
  const appId = useAppId();
  const guilds = useAppStateGuilds();

  const createMutation = useCreditLimitCreateMutation(appId);
  const updateMutation = useCreditLimitUpdateMutation(appId, limit?.id ?? "");
  const pending = createMutation.isPending || updateMutation.isPending;

  const form = useForm<FormFields>({
    defaultValues: defaultValues(limit, defaults),
  });

  const scope = form.watch("scope");
  const target = form.watch("target");
  const unlimited = form.watch("unlimited");
  const targetName = scope === "guild" ? "server" : "user";

  function onSubmit(data: FormFields) {
    if (pending) return;

    const specific = data.target === "specific";
    const unlimited = specific && data.unlimited;

    let credits: number | null = null;
    if (!unlimited) {
      credits = Number(data.credits);
      if (data.credits.trim() === "" || !Number.isInteger(credits)) {
        form.setError("credits", { message: "Enter a whole number" });
        return;
      }
    }

    const req = {
      scope: data.scope,
      target_id: specific ? data.target_id.trim() : null,
      period: data.period,
      credits,
    };

    const handlers = {
      onSuccess(res: Awaited<ReturnType<typeof createMutation.mutateAsync>>) {
        if (res.success) {
          toast.success(limit ? "Limit updated!" : "Limit created!");
          setOpen(false);
          if (!limit) form.reset(defaultValues(undefined, defaults));
        } else if (res.error.code === "validation_failed") {
          setValidationErrors(form, res.error.data);
        } else {
          toast.error(
            `Failed to save limit: ${res.error.message} (${res.error.code})`
          );
        }
      },
    };

    if (limit) {
      updateMutation.mutate(req, handlers);
    } else {
      createMutation.mutate(req, handlers);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (open) form.reset(defaultValues(limit, defaults));
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {limit ? "Edit Credit Limit" : "Create Credit Limit"}
          </DialogTitle>
          <DialogDescription>
            Once a {targetName} has used this many credits, commands, event
            listeners and buttons stop running for it until the{" "}
            {form.watch("period") === "day" ? "day" : "month"} ends (UTC).
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="scope"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Applies to</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="guild">Servers</SelectItem>
                        <SelectItem value="user">Users</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="period"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Resets</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="day">Daily</SelectItem>
                        <SelectItem value="month">Monthly</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name="target"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Target</FormLabel>
                  <Select value={field.value} onValueChange={field.onChange}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="default">
                        Every {targetName} (default)
                      </SelectItem>
                      <SelectItem value="specific">
                        A specific {targetName}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {target === "default"
                      ? `Applies to every ${targetName} without a limit of its own.`
                      : `Replaces the default limit for this ${targetName}.`}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            {target === "specific" && (
              <FormField
                control={form.control}
                name="target_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {scope === "guild" ? "Server ID" : "User ID"}
                    </FormLabel>
                    {scope === "guild" && guilds && guilds.length > 0 && (
                      <Select
                        value={
                          guilds.some((g) => g!.id === field.value)
                            ? field.value
                            : ""
                        }
                        onValueChange={field.onChange}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pick a server" />
                        </SelectTrigger>
                        <SelectContent>
                          {guilds.map((guild) => (
                            <SelectItem key={guild!.id} value={guild!.id}>
                              {guild!.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                    <FormControl>
                      <Input
                        type="text"
                        inputMode="numeric"
                        placeholder="123456789012345678"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            {target === "specific" && (
              <FormField
                control={form.control}
                name="unlimited"
                render={({ field }) => (
                  <FormItem className="flex items-center justify-between space-y-0 rounded-lg border p-3">
                    <div className="space-y-0.5">
                      <FormLabel>No limit</FormLabel>
                      <FormDescription>
                        Exempt this {targetName} from the default limit.
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            )}
            {(target === "default" || !unlimited) && (
              <FormField
                control={form.control}
                name="credits"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Credits</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={0}
                        step={1}
                        placeholder="1000"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      0 blocks the {targetName} completely.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <DialogFooter>
              <LoadingButton type="submit" loading={pending}>
                {limit ? "Save limit" : "Create limit"}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
