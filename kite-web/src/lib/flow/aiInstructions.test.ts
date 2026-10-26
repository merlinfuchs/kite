import { expect, test } from "vitest";
import { flowAIInstructions } from "./aiInstructions";

test("writes the flow AI instructions for the service", async () => {
  await expect(flowAIInstructions + "\n").toMatchFileSnapshot(
    "../../../../kite-service/internal/core/flowai/instructions.txt"
  );
});
