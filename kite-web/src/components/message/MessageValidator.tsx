import { Message, messageSchema } from "@/lib/message/schema";
import { useCurrentMessage } from "@/lib/message/state";
import { useValidationErrors } from "@/lib/message/state";
import debounce from "just-debounce-it";

export default function MessageValidator() {
  const setValidationError = useValidationErrors((state) => state.setError);

  const debouncedSetValidationError = debounce((msg: Message) => {
    const res = messageSchema.safeParse(msg);
    setValidationError(res.success ? null : res.error);
  }, 250);

  useCurrentMessage((state) => {
    debouncedSetValidationError(state);
    return null;
  });

  return null;
}
