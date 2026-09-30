import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiQrCodeCreate: BlockDefinition = {
  type: "action_cookie_api_qr_code_create",
  title: "Generate QR code",
  description: "Generate an image of a QR code with Cookie API",
  icon: "qr-code",
  category: "Cookie API",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_qr_code",
    method: "POST",
    path: "/api/images/qr-code",
  },
  fields: [
    {
      name: "qr_code_data",
      in: "body",
      target: "data",
      type: "string",
      label: "Content",
      description: "What the QR code contains, like a link.",
      required: true,
    },
    {
      name: "qr_code_style",
      in: "body",
      target: "style",
      type: "string",
      label: "Style",
      description: 'Either "dots" or "standart". Defaults to "standart".',
    },
    {
      name: "qr_code_border",
      in: "body",
      target: "border",
      type: "integer",
      label: "Border",
      description: "Size of the border around the QR code.",
      min: 0,
    },
    {
      name: "qr_code_center_image_url",
      in: "body",
      target: "center_image_url",
      type: "string",
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
