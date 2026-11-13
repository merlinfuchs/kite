import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiQrCodeCreate: BlockDefinition = {
  type: "action_cookie_api_qr_code_create",
  id_prefix: "qr",
  title: "Generate QR code",
  description: "Generate an image of a QR code with Cookie API",
  icon: "qr-code",
  category: "Cookie API",
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_qr_code",
  },
  fields: [
    {
      name: "qr_code_data",
      target: "data",
      label: "Content",
      description: "What the QR code contains, like a link.",
    },
    {
      name: "qr_code_style",
      target: "style",
      label: "Style",
      description: "How the QR code is drawn. Defaults to squares.",
      // "standart" is how Cookie API spells it.
      options: [
        { value: "standart", label: "Squares" },
        { value: "dots", label: "Dots" },
      ],
    },
    {
      name: "qr_code_border",
      target: "border",
      label: "Border",
      description: "Size of the border around the QR code.",
      min: 0,
    },
    {
      name: "qr_code_center_image_url",
      target: "center_image_url",
      label: "Center Image",
      description: "URL of an image to show in the middle of the QR code.",
    },
  ],
  result: {
    schema: z
      .object({
        url: z.string().describe("URL of the QR code's image"),
      })
      .describe("The generated QR code"),
  },
};
