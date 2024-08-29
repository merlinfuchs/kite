import { useValidationErrors } from "@/lib/message/state";
import BaseInput, { BaseInputProps } from "@/tools/common/components/BaseInput";

type Props = BaseInputProps & {
  validationPath?: string;
};

export default function MessageInput(props: Props) {
  const issue = useValidationErrors(
    (state) =>
      props.validationPath &&
      state.getIssueByPath(props.validationPath)?.message
  );

  return <BaseInput {...props} error={issue} />;
}
