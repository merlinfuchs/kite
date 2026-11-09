import { z } from "zod";
import { HandlerOperationDefinition } from "../handlerOperations";

export const httpOperations: HandlerOperationDefinition[] = [
  {
    id: "request",
    description:
      "Send an HTTP request to a URL, like the Send API Request block. {{secrets.NAME}} in the URL, the headers and the body is replaced with the app's secret, which the code never sees. Responses that aren't 2xx resolve too, only failing to get a response rejects. fetch(url, init) calls it.",
    credits: 3,
    params: z
      .object({
        url: z.string().describe("The URL, starting with http:// or https://."),
        method: z
          .enum(["GET", "POST", "PUT", "PATCH", "DELETE"])
          .optional()
          .describe("Defaults to GET."),
        headers: z
          .record(z.string(), z.string())
          .optional()
          .describe("Headers by name."),
        body: z
          .string()
          .optional()
          .describe("The body as text, like JSON.stringify(data)."),
      })
      .strict(),
    result: z.object({
      status: z.number().describe("The HTTP status, like 200."),
      status_text: z.string().describe("Like OK."),
      headers: z
        .record(z.string(), z.string())
        .describe(
          "Headers by name in lower case, several values joined by commas."
        ),
      body: z.string().describe("The body as text, up to 4 MB."),
    }),
  },
];
