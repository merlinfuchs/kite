import { useValidationErrors } from "@/lib/message/state";
import { ValidationScope } from "@/lib/message/validationStore";
import { CircleAlertIcon } from "lucide-react";

interface Props {
  scope: ValidationScope;
}

export default function MessageValidationErrorIndicator({ scope }: Props) {
  const error = useValidationErrors((state) => state.hasIssue(scope));

  if (error) {
    return (
      <CircleAlertIcon className="h-5 w-5 text-red-600 dark:text-red-400" />
    );
  } else {
    return null;
  }
}
