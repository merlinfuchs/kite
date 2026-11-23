import { z } from "zod";
import clientEnv from "./client";

export const serverEnvSchema = z.object({
  // Where the server reaches the API, like nginx on the same machine, so its
  // requests don't leave the machine. Defaults to the public URL.
  API_INTERNAL_BASE_URL: z.string().optional(),
});

const serverEnv = {
  ...serverEnvSchema.parse({
    API_INTERNAL_BASE_URL: process.env.API_INTERNAL_BASE_URL,
  }),
  ...clientEnv,
};

export default serverEnv;
