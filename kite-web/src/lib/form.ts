import { UseFormReturn } from "react-hook-form";

export function setValidationErrors(
  form: UseFormReturn<any, any, any>,
  errors: Record<string, any>
) {
  for (const [path, message] of Object.entries(errors)) {
    form.setError(path, {
      type: "server",
      message: message,
    });
  }
}
