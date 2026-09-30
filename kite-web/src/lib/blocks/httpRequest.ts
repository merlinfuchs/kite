import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { httpResponseResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const httpRequest: BlockDefinition = {
  type: "action_http_request",
  title: "Send API Request",
  description: "Send an API request to an external server",
  icon: "webhook",
  category: "API Requests",
  credits: 3,
  fields: [
    {
      name: "http_request_data",
      type: "string",
      schema: z
        .object({
          url: templated(z.string().url(), "URL to send the request to."),
          method: z
            .enum(["GET", "POST", "PUT", "PATCH", "DELETE"])
            .describe("HTTP method of the request."),
          headers: z
            .array(
              z.object({
                key: z.string().describe("Name of the header."),
                value: templated(z.string(), "Value of the header."),
              })
            )
            .optional()
            .describe("Headers to send with the request."),
          body_json: z
            .record(z.unknown())
            .optional()
            .describe(
              "JSON body of the request. Placeholders in its string values are evaluated."
            ),
        })
        .describe("The request to send."),
    },
  ],
  result: { schema: httpResponseResultSchema },
  run: { kind: "custom" },
};
