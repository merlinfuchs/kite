import { useValidationErrors } from "@/lib/message/state";
import { ValidationTarget } from "@/lib/message/validationStore";
import BaseInput, { BaseInputProps } from "@/tools/common/components/BaseInput";
import { useCallback } from "react";
import MessagePlaceholderExplorer from "./MessagePlaceholderExplorer";

type Props = BaseInputProps & {
  validation?: ValidationTarget;
  placeholders?: boolean;
};

export default function MessageInput(props: Props) {
  const { validation, placeholders, ...inputProps } = props;

  const issue = useValidationErrors(
    (state) => validation && state.getIssue(validation)?.message
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
      <BaseInput {...(inputProps as BaseInputProps)} error={issue} />
      {placeholders && (
        <MessagePlaceholderExplorer onSelect={onPlaceholderSelect} />
      )}
    </div>
  );
}
