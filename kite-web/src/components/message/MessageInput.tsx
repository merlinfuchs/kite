import { useValidationErrors } from "@/lib/message/state";
import BaseInput, { BaseInputProps } from "@/tools/common/components/BaseInput";
import { useCallback } from "react";
import MessagePlaceholderExplorer from "./MessagePlaceholderExplorer";

type Props = BaseInputProps & {
  validationPath?: string;
  placeholders?: boolean;
};

export default function MessageInput(props: Props) {
  const issue = useValidationErrors(
    (state) =>
      props.validationPath &&
      state.getIssueByPath(props.validationPath)?.message
  );

  const onPlaceholderSelect = useCallback(
    (placeholder: string) => {
      const value = `{{${placeholder}}}`;

      // TODO?: This is pretty hacky, we should think about baking placeholder support into the BaseInput component
      props.onChange((props.value + value) as never);
    },
    [props]
  );

  return (
    <div className="relative w-full">
      <BaseInput {...props} error={issue} />
      {props.placeholders && (
        <MessagePlaceholderExplorer onSelect={onPlaceholderSelect} />
      )}
    </div>
  );
}
