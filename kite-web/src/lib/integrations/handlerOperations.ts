import { z } from "zod";
import { aiOperations } from "./ai/operations";
import { httpOperations } from "./http/operations";
import { variablesOperations } from "./variables/operations";

// An operation that runs Go code of the service instead of sending a
// request, like http.request or the stored variables. The service runs it
// with the handler registered under its name in operationHandlers. See
// "Operation kinds" in design/integrations.md.
export interface HandlerOperationDefinition {
  // Name of the operation in its group, in snake case like operationIds.
  id: string;
  description: string;
  // What a call costs, or the least it costs with variable_credits.
  credits: number;
  // The credits depend on the argument, which the description explains.
  variable_credits?: boolean;
  // The argument code passes, an object.
  params: z.ZodType;
  result: z.ZodType;
  // It deletes something, see OperationPolicy.
  destructive?: boolean;
}

// The operations of Kite's own features, by group. Code calls them like
// those of integrations, as http.request(arg).
export const handlerOperations: Record<string, HandlerOperationDefinition[]> = {
  http: httpOperations,
  variables: variablesOperations,
  ai: aiOperations,
};
