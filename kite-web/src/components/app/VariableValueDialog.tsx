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
import { useVariableValueSetMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { useAppId } from "@/lib/hooks/params";
import { Variable, VariableValue } from "@/lib/types/wire.gen";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Textarea } from "../ui/textarea";

interface FormFields {
  scope: string;
  type: string;
  value: string;
}

export const variableValueTypes = [
  { value: "string", label: "Text" },
  { value: "number", label: "Number" },
  { value: "bool", label: "True / False" },
  { value: "json", label: "JSON" },
];

function defaultValues(value?: VariableValue): FormFields {
  return {
    scope: value?.scope ?? "",
    type: value?.type ?? "string",
    value: value?.value ?? "",
  };
}

// Values a flow stored that can't be entered as text, ones too large to load,
// and ones whose scope doesn't fit the variable, like a scoped value in an
// unscoped variable, can only be viewed.
export function isValueReadOnly(variable: Variable, value: VariableValue) {
  return (
    value.read_only ||
    value.truncated ||
    (value.scope !== null) !== variable.scoped
  );
}

// Adds a value to the variable, or edits one when given.
export default function VariableValueDialog({
  children,
  variable,
  value,
}: {
  children: ReactNode;
  variable: Variable;
  value?: VariableValue;
}) {
  const [open, setOpen] = useState(false);

  const setMutation = useVariableValueSetMutation(useAppId(), variable.id);
  const readOnly = !!value && isValueReadOnly(variable, value);
  const showScope = variable.scoped || value?.scope != null;

  const form = useForm<FormFields>({
    defaultValues: defaultValues(value),
  });
  const type = form.watch("type");

  function onSubmit(data: FormFields) {
    if (setMutation.isPending || readOnly) return;

    setMutation.mutate(
      {
        // The scope of an existing value can't be changed.
        scope: value ? value.scope ?? "" : data.scope,
        type: data.type,
        value: data.value,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Value saved!");
            setOpen(false);
          } else if (res.error.code === "validation_failed") {
            setValidationErrors(form, res.error.data);
          } else {
            toast.error(
              `Failed to save value: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        // Start from the stored value each time, it may have changed since.
        if (open) form.reset(defaultValues(value));
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {readOnly ? "View Value" : value ? "Edit Value" : "Add Value"}
          </DialogTitle>
          <DialogDescription>
            {readOnly
              ? value?.truncated
                ? "This value is too large to edit here, only the start of it is shown."
                : value?.read_only
                ? "This value holds data stored by a flow that can't be edited here."
                : variable.scoped
                ? "This value has no scope, so it can't be edited here."
                : "This value has a scope but the variable isn't scoped, so it can't be edited here."
              : variable.scoped && !value
              ? "If there already is a value for the scope, it will be replaced."
              : "Flows that run after saving will read the new value."}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            {showScope && (
              <FormField
                control={form.control}
                name="scope"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Scope</FormLabel>
                    <FormControl>
                      <Input
                        type="text"
                        placeholder={value ? "No scope" : "e.g. a user ID"}
                        disabled={!!value}
                        {...field}
                      />
                    </FormControl>
                    {!value && (
                      <FormDescription>
                        The key the value is stored under, like a user, channel
                        or server ID.
                      </FormDescription>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select
                    value={field.value}
                    onValueChange={(type) => {
                      field.onChange(type);
                      form.clearErrors("value");
                      // The select below only knows these two values.
                      if (type === "bool") {
                        const current = form.getValues("value").trim();
                        form.setValue(
                          "value",
                          current === "true" ? "true" : "false"
                        );
                      }
                    }}
                    disabled={readOnly}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {variableValueTypes.map((t) => (
                        <SelectItem value={t.value} key={t.value}>
                          {t.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Value</FormLabel>
                  {type === "bool" ? (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={readOnly}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="true">True</SelectItem>
                        <SelectItem value="false">False</SelectItem>
                      </SelectContent>
                    </Select>
                  ) : type === "number" ? (
                    <FormControl>
                      {/* Not type="number", which rounds numbers as large as Discord IDs. */}
                      <Input
                        type="text"
                        inputMode="decimal"
                        placeholder="0"
                        readOnly={readOnly}
                        {...field}
                      />
                    </FormControl>
                  ) : (
                    <FormControl>
                      <Textarea
                        rows={type === "json" ? 8 : 4}
                        className={type === "json" ? "font-mono text-xs" : ""}
                        placeholder={
                          type === "json" ? '{"key": "value"}' : undefined
                        }
                        readOnly={readOnly}
                        {...field}
                      />
                    </FormControl>
                  )}
                  {type === "json" && !readOnly && (
                    <FormDescription>
                      A list, an object or <code>null</code>.
                    </FormDescription>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />
            {!readOnly && (
              <DialogFooter>
                <LoadingButton type="submit" loading={setMutation.isPending}>
                  Save value
                </LoadingButton>
              </DialogFooter>
            )}
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
