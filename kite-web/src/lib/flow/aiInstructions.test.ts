import { expect, test } from "vitest";
import { flowAICodeInstructions, flowAIInstructions } from "./aiInstructions";

test("writes the flow AI instructions for the service", async () => {
  await expect(flowAIInstructions + "\n").toMatchFileSnapshot(
    "../../../../kite-service/internal/core/flowai/instructions.txt"
  );
});

test("writes the flow AI instructions for code for the service", async () => {
  await expect(flowAICodeInstructions + "\n").toMatchFileSnapshot(
    "../../../../kite-service/internal/core/flowai/instructions_code.txt"
  );
});
