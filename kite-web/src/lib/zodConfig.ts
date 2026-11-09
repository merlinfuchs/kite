import { z } from "zod";

// Keeps zod 3's message for missing values, which the flow AI and the message
// editor show. Other issues use zod 4's messages.
z.config({
  customError: (issue) => (issue.input === undefined ? "Required" : undefined),
});
