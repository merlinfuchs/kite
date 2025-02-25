import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "../ui/input";
import { useForm } from "react-hook-form";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { useCallback, useEffect } from "react";
import { useVariable } from "@/lib/hooks/api";
import { useVariableUpdateMutation } from "@/lib/api/mutations";
import { useAppId, useVariableId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import ConfirmDialog from "../common/ConfirmDialog";
import { Switch } from "../ui/switch";

interface FormFields {
  name: string;
  scoped: boolean;
}

export default function VariableSettingsCore() {
  const variable = useVariable();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      scoped: false,
    },
  });

  useEffect(() => {
    if (variable) {
      form.reset({
        name: variable.name,
        scoped: variable.scoped,
      });
    }
  }, [variable, form]);

  const updateMutation = useVariableUpdateMutation(useAppId(), useVariableId());

  const saveSettings = useCallback(() => {
    const data = form.getValues();

    updateMutation.mutate(
      {
        name: data.name,
        scoped: data.scoped,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Settings saved!");
          } else {
            if (res.error.code === "validation_failed") {
              setValidationErrors(form, res.error.data);
            } else {
              toast.error(
                `Failed to update app: ${res.error.message} (${res.error.code})`
              );
            }
          }
        },
      }
    );
  }, [form, updateMutation]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Variable Settings</CardTitle>
        <CardDescription>
          Configure how your app appears to users in Discord and Kite.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form className="grid gap-4">
          <CardContent className="space-y-5">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="scoped"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center justify-between rounded-lg border px-4 py-3">
                  <div className="space-y-0.5">
                    <FormLabel className="text-sm">Scoped</FormLabel>
                    <FormDescription>
                      A scoped variables allows storing multiple values scoped
                      by a specific key. This can be useful for storing
                      user-specfic data.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      aria-readonly
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="border-t px-6 py-4">
            <ConfirmDialog
              title="Are you sure that you want to update the variable settings?"
              description="Changing if the variable is scoped or not will delete all associated data and cannot be undone."
              onConfirm={saveSettings}
            >
              <Button>Save settings</Button>
            </ConfirmDialog>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
