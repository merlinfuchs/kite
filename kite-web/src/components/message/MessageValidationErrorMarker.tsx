import { useValidationErrors } from "@/lib/message/state";
import { CircleAlertIcon } from "lucide-react";

interface Props {
  pathPrefix: string | string[];
}

export default function MessageValidationErrorIndicator({ pathPrefix }: Props) {
  const error = useValidationErrors((state) =>
    typeof pathPrefix === "string"
      ? state.checkIssueByPathPrefix(pathPrefix)
      : pathPrefix.some((prefix) => state.checkIssueByPathPrefix(prefix))
  );

  if (error) {
    return (
      <CircleAlertIcon className="h-5 w-5 text-red-600 dark:text-red-400" />
    );
  } else {
    return null;
  }
}
