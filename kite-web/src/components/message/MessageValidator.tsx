import { messageSchema } from "@/lib/message/schema";
import { useDocumentStoreApi, useValidationErrors } from "@/lib/message/state";
import { toMessage } from "@/lib/message/documentConvert";
import debounce from "just-debounce-it";
import { useEffect } from "react";

export default function MessageValidator() {
  const store = useDocumentStoreApi();
  const setValidationError = useValidationErrors((state) => state.setError);

  useEffect(() => {
    const validate = debounce(() => {
      const { message, idToPath } = toMessage(store.getState());
      const res = messageSchema.safeParse(message);
      setValidationError(res.success ? null : res.error, idToPath);
    }, 250);

    validate();
    return store.subscribe(validate);
  }, [store, setValidationError]);

  return null;
}
