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
import { useApp, useVariable } from "@/lib/hooks/api";
import {
  useAppUpdateMutation,
  useVariableUpdateMutation,
} from "@/lib/api/mutations";
import { useAppId, useVariableId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { variableScopes, variableTypes } from "@/lib/variable";

interface FormFields {
  name: string;
  scope: string;
  type: string;
}

export default function VariableSettingsCore() {
  const variable = useVariable();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      scope: "",
      type: "",
    },
  });

  useEffect(() => {
    if (variable) {
      form.reset({
        name: variable.name,
        scope: variable.scope,
        type: variable.type,
      });
    }
  }, [variable, form]);

  const updateMutation = useVariableUpdateMutation(useAppId(), useVariableId());

  const onSubmit = useCallback(
    (data: FormFields) => {
      updateMutation.mutate(
        {
          name: data.name,
          scope: data.scope,
          type: data.type,
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
    },
    [form, updateMutation]
  );

  return (
    <Card x-chunk="dashboard-04-chunk-1">
      <CardHeader>
        <CardTitle>Variable Settings</CardTitle>
        <CardDescription>
          Configure how your app appears to users in Discord and Kite.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
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
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select the type of the variable" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {variableTypes.map((type) => (
                        <SelectItem value={type.value} key={type.value}>
                          {type.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {
                      variableTypes.find((type) => type.value === field.value)
                        ?.description
                    }
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="scope"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Scope</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select the scope for the variable" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {variableScopes.map((scope) => (
                        <SelectItem value={scope.value} key={scope.value}>
                          {scope.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {
                      variableScopes.find(
                        (scope) => scope.value === field.value
                      )?.description
                    }
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="border-t px-6 py-4">
            <Button type="submit">Save settings</Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}
