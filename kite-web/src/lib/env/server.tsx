import { z } from "zod";
import clientEnv from "./client";

export const serverEnvSchema = z.object({});

const serverEnv = {
  ...serverEnvSchema.parse({}),
  ...clientEnv,
};

export default serverEnv;
