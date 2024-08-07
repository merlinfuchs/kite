import { Message, messageSchema } from "../schema/message";
import { useCurrentMessageStore } from "../state/message";
import { useValidationErrorStore } from "../state/validationError";
import debounce from "just-debounce-it";

export default function MessageValidator() {
  const setValidationError = useValidationErrorStore((state) => state.setError);

  const debouncedSetValidationError = debounce((msg: Message) => {
    const res = messageSchema.safeParse(msg);
    setValidationError(res.success ? null : res.error);
  }, 250);

  useCurrentMessageStore((state) => {
    debouncedSetValidationError(state);
    return null;
  });

  return null;
}
