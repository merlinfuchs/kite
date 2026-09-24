import { useValidationErrors } from "@/lib/message/state";
import { ValidationTarget } from "@/lib/message/validationStore";
import { CircleAlertIcon } from "lucide-react";

interface Props {
  target: ValidationTarget;
}

export default function MessageValidationError({ target }: Props) {
  const issue = useValidationErrors((state) => state.getIssue(target)?.message);

  if (issue) {
    return (
      <div className="text-red-600 dark:text-red-400 text-sm flex items-center space-x-1 pt-1">
        <CircleAlertIcon className="h-5 w-5 flex-none" />
        <div>{issue}</div>
      </div>
    );
  } else {
    return null;
  }
}
