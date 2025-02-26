import { UseFormReturn } from "react-hook-form";

export function setValidationErrors(
  form: UseFormReturn<any, any, any>,
  errors: Record<string, any>,
  aliases: Record<string, string> = {}
) {
  function walkErrors(
    obj: any,
    path: string[] = [],
    result: Record<string, string> = {}
  ) {
    for (const [key, value] of Object.entries(obj)) {
      const currentPath = [...path, key];

      if (typeof value === "string") {
        result[currentPath.join(".")] = value;
      } else if (typeof value === "object" && value !== null) {
        walkErrors(value, currentPath, result);
      }
    }
    return result;
  }

  const flatErrors = walkErrors(errors);

  for (const [path, message] of Object.entries(flatErrors)) {
    const realPath = aliases[path] || path;

    form.setError(realPath, {
      message,
    });
  }
}
