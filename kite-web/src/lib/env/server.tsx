import { z } from "zod";
import clientEnv from "./client";

export const serverEnvSchema = z.object({});

export default {
  ...serverEnvSchema.parse({}),
  ...clientEnv,
};
